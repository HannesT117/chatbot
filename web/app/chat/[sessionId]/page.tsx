"use client";

import { useParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { useChat } from "./useChat";
import { ChatMessage } from "@/components/ChatMessage";
import { StreamingMessage } from "@/components/StreamingMessage";
import { ChatInput } from "@/components/ChatInput";

export default function ChatPage() {
  const params = useParams<{ sessionId: string }>();
  const { messages, streaming, provisionalText, send } = useChat(params.sessionId);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, provisionalText]);

  return (
    <div className="flex flex-col h-screen max-w-2xl mx-auto">
      <header className="p-4 border-b">
        <h1 className="font-semibold">Chat</h1>
        <p className="text-xs text-muted-foreground">Session: {params.sessionId}</p>
      </header>

      <div className="flex-1 overflow-y-auto p-4">
        {messages.map((msg) => (
          <ChatMessage key={msg.id} message={msg} />
        ))}
        {streaming && <StreamingMessage text={provisionalText} />}
        <div ref={bottomRef} />
      </div>

      <ChatInput onSend={send} disabled={streaming} />
    </div>
  );
}
