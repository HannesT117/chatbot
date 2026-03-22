import { useState, useCallback } from "react";
import type { Message } from "@/components/ChatMessage";

type UseChatReturn = {
  messages: Message[];
  streaming: boolean;
  provisionalText: string;
  send: (text: string) => void;
};

export function useChat(sessionId: string): UseChatReturn {
  const [messages, setMessages] = useState<Message[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [provisionalText, setProvisionalText] = useState("");

  const send = useCallback(
    async (text: string) => {
      if (streaming) return;

      setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: "user", text }]);
      setStreaming(true);
      setProvisionalText("");

      let accumulated = "";
      let settled = false;

      function settle(msg: Message) {
        setMessages((prev) => [...prev, msg]);
        setProvisionalText("");
        setStreaming(false);
        accumulated = "";
        settled = true;
      }

      try {
        const res = await fetch("/api/chat", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ session_id: sessionId, message: text }),
        });

        if (!res.ok || !res.body) {
          throw new Error(`HTTP ${res.status}`);
        }

        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const rawEvents = buffer.split("\n\n");
          buffer = rawEvents.pop() ?? "";

          for (const raw of rawEvents) {
            const lines = raw.trim().split("\n");
            const eventLine = lines.find((l) => l.startsWith("event:"));
            const dataLine = lines.find((l) => l.startsWith("data:"));
            if (!eventLine || !dataLine) continue;

            const event = eventLine.slice("event:".length).trim();
            const data = JSON.parse(dataLine.slice("data:".length).trim());

            if (event === "token") {
              accumulated += data.text;
              setProvisionalText(accumulated);
            } else if (event === "done") {
              settle({ id: crypto.randomUUID(), role: "assistant", text: accumulated });
            } else if (event === "blocked") {
              settle({ id: crypto.randomUUID(), role: "blocked", text: "" });
            } else if (event === "error") {
              settle({ id: crypto.randomUUID(), role: "error", text: data.message ?? "An error occurred" });
            }
          }
        }

        if (!settled) {
          settle({ id: crypto.randomUUID(), role: "error", text: "Connection lost" });
        }
      } catch (err) {
        settle({
          id: crypto.randomUUID(),
          role: "error",
          text: err instanceof Error ? err.message : "An error occurred",
        });
      }
    },
    [sessionId, streaming]
  );

  return { messages, streaming, provisionalText, send };
}
