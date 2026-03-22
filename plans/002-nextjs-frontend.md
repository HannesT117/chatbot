# Next.js Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Next.js presentation layer that consumes the Go server's REST+SSE API — scenario picker, chat UI with real-time streaming, and clean blocked-message handling.

**Architecture:** A Next.js App Router app in `web/`. The home page fetches available scenarios and creates a session; the chat page streams LLM responses token-by-token via SSE. All Go server communication is typed fetch calls; no LLM keys or session internals exist in the frontend.

**Tech Stack:** Next.js 15 (App Router), React, TypeScript, Tailwind CSS, shadcn/ui (`Card`, `Button`, `Textarea`, `Badge`, `Alert`), React Testing Library, `@testing-library/user-event`.

---

## SSE Event Contract

The Go server emits named SSE events over the `POST /api/chat` response. This frontend is built against this contract — the Go server must match it exactly.

```
event: token
data: {"text":"Hello"}

event: blocked
data: {"reason":"output_blocklist"}

event: done
data: {}

event: error
data: {"message":"internal server error"}
```

- `token` — append `text` to the in-progress assistant message
- `done` — message complete, commit provisional text to message history
- `blocked` — discard provisional tokens, insert a blocked notice in history
- `error` — discard provisional tokens, insert an error notice in history

## Go Server API

| Method | Path | Request | Response |
|--------|------|---------|----------|
| `GET` | `/api/scenarios` | — | `{"scenarios":[{"name":"financial_advisor","persona_name":"Morgan","persona_description":"..."},...]}` |
| `POST` | `/api/sessions` | `{"scenario_name":"financial_advisor"}` | `{"session_id":"<uuid>"}` |
| `POST` | `/api/chat` | `{"session_id":"<uuid>","message":"<text>"}` | SSE stream (see above) |
| `DELETE` | `/api/sessions/:id` | — | 204 |

The Go server runs on `http://localhost:8080` in development. Set `API_URL` (server-side only, not inlined into the client bundle) to override. Browser-origin requests go through the Next.js rewrite proxy at `/api/*`.

## File Map

```
web/
  AGENTS.md                              # Next.js agent rules (new)
  CLAUDE.md                              # @AGENTS.md import (new)
  package.json                           # Next.js, Tailwind, shadcn, RTL (new)
  next.config.ts                         # Rewrite proxy for browser-origin API calls (new)
  tsconfig.json                          # TS config (new)
  tailwind.config.ts                     # Tailwind config (new)
  components.json                        # shadcn config (new)
  app/
    layout.tsx                           # Root layout, font, global styles (new)
    page.tsx                             # Scenario picker (new)
    error.tsx                            # Error boundary: shown when Go server is unreachable (new)
    chat/
      [sessionId]/
        page.tsx                         # Chat UI page (new)
        useChat.ts                       # SSE hook + Message type (new)
  components/
    ScenarioCard.tsx                     # Scenario card — pure display, no interactivity (new)
    ChatMessage.tsx                      # Single message: user / assistant / blocked / error (new)
    StreamingMessage.tsx                 # In-progress assistant message + blinking cursor (new)
    ChatInput.tsx                        # Textarea + send button, disabled while streaming (new)
  __tests__/
    ScenarioCard.test.tsx
    ChatMessage.test.tsx
    StreamingMessage.test.tsx
    ChatInput.test.tsx
    useChat.test.ts
```

---

## Task 1: Scaffold the Next.js project

**Files:**
- Create: `web/package.json`
- Create: `web/next.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/tailwind.config.ts`
- Create: `web/components.json`
- Create: `web/AGENTS.md`
- Create: `web/CLAUDE.md`
- Create: `web/app/layout.tsx`

- [ ] **Step 1: Initialise the project**

```bash
cd /path/to/chatbot
npx create-next-app@latest web \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --no-src-dir \
  --import-alias "@/*" \
  --no-agents-md
```

When prompted for any additional options, accept the defaults.

- [ ] **Step 2: Install shadcn/ui**

```bash
cd web
npx shadcn@latest init
```

Accept all defaults (New York style, zinc base colour, CSS variables: yes).

Then add the components we need:

```bash
npx shadcn@latest add card button textarea badge alert
```

- [ ] **Step 3: Create AGENTS.md and CLAUDE.md**

`web/AGENTS.md`:
```md
<!-- BEGIN:nextjs-agent-rules -->

# Next.js: ALWAYS read docs before coding

Before any Next.js work, find and read the relevant doc in `node_modules/next/dist/docs/`. Your training data is outdated — the docs are the source of truth.

<!-- END:nextjs-agent-rules -->
```

`web/CLAUDE.md`:
```md
@AGENTS.md
```

- [ ] **Step 4: Configure the API proxy in next.config.ts**

Replace the generated `next.config.ts` with:

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    const apiUrl = process.env.API_URL ?? "http://localhost:8080";
    return [
      {
        source: "/api/:path*",
        destination: `${apiUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
```

This proxies browser-origin `/api/*` requests to the Go server, so the browser never speaks directly to the Go server (no CORS configuration needed). Server-side fetches (RSC, server actions) call the Go server directly using `API_URL`.

- [ ] **Step 5: Write error boundary**

`web/app/error.tsx`:
```typescript
"use client";

export default function Error({ error }: { error: Error }) {
  return (
    <main className="flex min-h-screen items-center justify-center p-8">
      <div className="text-center">
        <h1 className="text-xl font-semibold mb-2">Something went wrong</h1>
        <p className="text-muted-foreground text-sm">{error.message}</p>
        <p className="text-muted-foreground text-xs mt-1">Is the Go server running?</p>
      </div>
    </main>
  );
}
```

- [ ] **Step 6: Write root layout**

`web/app/layout.tsx`:
```typescript
import type { Metadata } from "next";
import { Geist } from "next/font/google";
import "./globals.css";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Guardrailed Chatbot",
  description: "LLM guardrails testbed",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${geist.className} bg-background text-foreground min-h-screen`}>
        {children}
      </body>
    </html>
  );
}
```

- [ ] **Step 7: Verify the scaffold builds**

```bash
cd web && npm run build
```

Expected: build succeeds with no TypeScript or lint errors.

- [ ] **Step 8: Commit**

```bash
cd web
git add AGENTS.md CLAUDE.md next.config.ts app/layout.tsx app/error.tsx app/globals.css \
        components.json tailwind.config.ts tsconfig.json package.json package-lock.json
git commit -m "Add Next.js project scaffold with shadcn/ui and API proxy"
```

---

## Task 2: ScenarioCard component

**Files:**
- Create: `web/components/ScenarioCard.tsx`
- Create: `web/__tests__/ScenarioCard.test.tsx`

- [ ] **Step 1: Install test dependencies**

```bash
cd web
npm install --save-dev @testing-library/react @testing-library/user-event @testing-library/jest-dom jest jest-environment-jsdom ts-jest
```

Add to `web/package.json` scripts:
```json
"test": "jest",
"test:watch": "jest --watch"
```

Add `web/jest.config.ts`:
```typescript
import type { Config } from "jest";

const config: Config = {
  testEnvironment: "jsdom",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/$1",
    "\\.(css|less|scss|sass)$": "identity-obj-proxy",
  },
  transform: {
    "^.+\\.tsx?$": ["ts-jest", { tsconfig: { jsx: "react-jsx" } }],
  },
};

export default config;
```

Add `web/jest.setup.ts`:
```typescript
import "@testing-library/jest-dom";
```

```bash
npm install --save-dev identity-obj-proxy
```

- [ ] **Step 2: Write the failing test**

`web/__tests__/ScenarioCard.test.tsx`:
```typescript
import { render, screen } from "@testing-library/react";
import { ScenarioCard, type Scenario } from "@/components/ScenarioCard";

const scenario: Scenario = {
  name: "financial_advisor",
  persona_name: "Morgan",
  persona_description: "A cautious financial literacy assistant.",
};

describe("ScenarioCard", () => {
  it("renders persona name and description", () => {
    render(<ScenarioCard scenario={scenario} />);
    expect(screen.getByText("Morgan")).toBeInTheDocument();
    expect(screen.getByText(/cautious financial/)).toBeInTheDocument();
  });
});
```

- [ ] **Step 3: Run test — expect FAIL**

```bash
cd web && npm test -- ScenarioCard
```

Expected: `Cannot find module '@/components/ScenarioCard'`

- [ ] **Step 4: Implement ScenarioCard**

`ScenarioCard` is a pure display component — no buttons, no interactivity. The caller decides how to wrap it (form submit on the home page, router.push on a client page if needed later).

`web/components/ScenarioCard.tsx`:
```typescript
import { Card, CardHeader, CardDescription } from "@/components/ui/card";

export type Scenario = {
  name: string;
  persona_name: string;
  persona_description: string;
};

type Props = {
  scenario: Scenario;
};

// ScenarioCard is rendered inside a <button>. Do NOT use heading elements
// (CardTitle renders <h3>) — headings inside buttons are invalid HTML.
export function ScenarioCard({ scenario }: Props) {
  return (
    <Card className="w-full hover:border-primary transition-colors">
      <CardHeader>
        <p className="font-semibold text-base">{scenario.persona_name}</p>
        <CardDescription>{scenario.persona_description}</CardDescription>
      </CardHeader>
    </Card>
  );
}
```

- [ ] **Step 5: Run test — expect PASS**

```bash
cd web && npm test -- ScenarioCard
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add components/ScenarioCard.tsx __tests__/ScenarioCard.test.tsx jest.config.ts jest.setup.ts
git commit -m "Add ScenarioCard component with tests"
```

---

## Task 3: ChatMessage and StreamingMessage components

**Files:**
- Create: `web/components/ChatMessage.tsx`
- Create: `web/components/StreamingMessage.tsx`
- Create: `web/__tests__/ChatMessage.test.tsx`
- Create: `web/__tests__/StreamingMessage.test.tsx`

The `Message` type belongs to `useChat.ts` (Task 5) — for now, define it here and re-export from there later. To avoid circular imports, define the type in a shared location accessible to both components and the hook. Given the small surface, inline it in `ChatMessage.tsx` and import it in `useChat.ts`.

- [ ] **Step 1: Write failing tests**

`web/__tests__/ChatMessage.test.tsx`:
```typescript
import { render, screen } from "@testing-library/react";
import { ChatMessage, type Message } from "@/components/ChatMessage";

describe("ChatMessage", () => {
  it("renders user message on the right", () => {
    const msg: Message = { id: "1", role: "user", text: "Hello" };
    render(<ChatMessage message={msg} />);
    expect(screen.getByText("Hello")).toBeInTheDocument();
    expect(screen.getByText("Hello").closest("[data-role='user']")).toBeTruthy();
  });

  it("renders assistant message on the left", () => {
    const msg: Message = { id: "2", role: "assistant", text: "Hi there" };
    render(<ChatMessage message={msg} />);
    expect(screen.getByText("Hi there")).toBeInTheDocument();
  });

  it("renders blocked message with alert", () => {
    const msg: Message = { id: "3", role: "blocked", text: "" };
    render(<ChatMessage message={msg} />);
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText(/response was blocked/i)).toBeInTheDocument();
  });

  it("renders error message with alert", () => {
    const msg: Message = { id: "4", role: "error", text: "Something went wrong" };
    render(<ChatMessage message={msg} />);
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });
});
```

`web/__tests__/StreamingMessage.test.tsx`:
```typescript
import { render, screen } from "@testing-library/react";
import { StreamingMessage } from "@/components/StreamingMessage";

describe("StreamingMessage", () => {
  it("renders partial text with a cursor", () => {
    render(<StreamingMessage text="Hello wor" />);
    expect(screen.getByText(/Hello wor/)).toBeInTheDocument();
    expect(document.querySelector("[aria-label='typing cursor']")).toBeInTheDocument();
  });

  it("renders empty string with just the cursor", () => {
    render(<StreamingMessage text="" />);
    expect(document.querySelector("[aria-label='typing cursor']")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
cd web && npm test -- ChatMessage StreamingMessage
```

- [ ] **Step 3: Implement ChatMessage**

`web/components/ChatMessage.tsx`:
```typescript
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { AlertTriangle } from "lucide-react";

export type Message = {
  id: string;
  role: "user" | "assistant" | "blocked" | "error";
  text: string;
};

type Props = {
  message: Message;
};

export function ChatMessage({ message }: Props) {
  if (message.role === "blocked") {
    return (
      <div className="flex justify-start my-2">
        <Alert variant="destructive" className="max-w-[80%]" role="alert">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            This response was blocked by the content filter.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (message.role === "error") {
    return (
      <div className="flex justify-start my-2">
        <Alert variant="destructive" className="max-w-[80%]" role="alert">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      </div>
    );
  }

  const isUser = message.role === "user";

  return (
    <div
      className={`flex my-2 ${isUser ? "justify-end" : "justify-start"}`}
      data-role={message.role}
    >
      <div
        className={`max-w-[80%] rounded-lg px-4 py-2 text-sm ${
          isUser
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-muted-foreground"
        }`}
      >
        {message.text}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Implement StreamingMessage**

`web/components/StreamingMessage.tsx`:
```typescript
type Props = {
  text: string;
};

export function StreamingMessage({ text }: Props) {
  return (
    <div className="flex justify-start my-2">
      <div className="max-w-[80%] rounded-lg px-4 py-2 text-sm bg-muted text-muted-foreground">
        {text}
        <span
          aria-label="typing cursor"
          className="inline-block w-0.5 h-4 bg-current ml-0.5 align-middle animate-pulse"
        />
      </div>
    </div>
  );
}
```

- [ ] **Step 5: Run tests — expect PASS**

```bash
cd web && npm test -- ChatMessage StreamingMessage
```

- [ ] **Step 6: Commit**

```bash
git add components/ChatMessage.tsx components/StreamingMessage.tsx \
        __tests__/ChatMessage.test.tsx __tests__/StreamingMessage.test.tsx
git commit -m "Add ChatMessage and StreamingMessage components with tests"
```

---

## Task 4: ChatInput component

**Files:**
- Create: `web/components/ChatInput.tsx`
- Create: `web/__tests__/ChatInput.test.tsx`

- [ ] **Step 1: Write failing test**

`web/__tests__/ChatInput.test.tsx`:
```typescript
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ChatInput } from "@/components/ChatInput";

describe("ChatInput", () => {
  it("calls onSend with the message text and clears the input", async () => {
    const user = userEvent.setup();
    const onSend = jest.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.type(screen.getByRole("textbox"), "Hello there");
    await user.click(screen.getByRole("button", { name: /send/i }));
    expect(onSend).toHaveBeenCalledWith("Hello there");
    expect(screen.getByRole("textbox")).toHaveValue("");
  });

  it("submits on Enter (without shift)", async () => {
    const user = userEvent.setup();
    const onSend = jest.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.type(screen.getByRole("textbox"), "Hello{Enter}");
    expect(onSend).toHaveBeenCalledWith("Hello");
  });

  it("does not submit on Shift+Enter", async () => {
    const user = userEvent.setup();
    const onSend = jest.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.type(screen.getByRole("textbox"), "Hello{Shift>}{Enter}{/Shift}");
    expect(onSend).not.toHaveBeenCalled();
  });

  it("disables input and button while streaming", () => {
    render(<ChatInput onSend={() => {}} disabled={true} />);
    expect(screen.getByRole("textbox")).toBeDisabled();
    expect(screen.getByRole("button", { name: /send/i })).toBeDisabled();
  });

  it("does not call onSend for blank input", async () => {
    const user = userEvent.setup();
    const onSend = jest.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.click(screen.getByRole("button", { name: /send/i }));
    expect(onSend).not.toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd web && npm test -- ChatInput
```

- [ ] **Step 3: Implement ChatInput**

`web/components/ChatInput.tsx`:
```typescript
"use client";

import { useState } from "react";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";

type Props = {
  onSend: (text: string) => void;
  disabled: boolean;
};

export function ChatInput({ onSend, disabled }: Props) {
  const [text, setText] = useState("");

  function submit() {
    const trimmed = text.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setText("");
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  }

  return (
    <div className="flex gap-2 items-end p-4 border-t">
      <Textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder="Type a message… (Enter to send, Shift+Enter for newline)"
        disabled={disabled}
        rows={2}
        className="resize-none flex-1"
      />
      <Button onClick={submit} disabled={disabled} aria-label="Send">
        <Send className="h-4 w-4" />
        <span className="sr-only">Send</span>
      </Button>
    </div>
  );
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
cd web && npm test -- ChatInput
```

- [ ] **Step 5: Commit**

```bash
git add components/ChatInput.tsx __tests__/ChatInput.test.tsx
git commit -m "Add ChatInput component with tests"
```

---

## Task 5: useChat hook (SSE streaming)

**Files:**
- Create: `web/app/chat/[sessionId]/useChat.ts`
- Create: `web/__tests__/useChat.test.ts`

- [ ] **Step 1: Write failing tests**

`web/__tests__/useChat.test.ts`:
```typescript
import { renderHook, act, waitFor } from "@testing-library/react";
import { useChat } from "@/app/chat/[sessionId]/useChat";

function makeSSEStream(events: Array<{ event: string; data: string }>) {
  const encoder = new TextEncoder();
  const chunks = events.map(({ event, data }) =>
    encoder.encode(`event: ${event}\ndata: ${data}\n\n`)
  );
  let i = 0;
  return new ReadableStream({
    pull(controller) {
      if (i < chunks.length) controller.enqueue(chunks[i++]);
      else controller.close();
    },
  });
}

global.fetch = jest.fn();

describe("useChat", () => {
  beforeEach(() => jest.clearAllMocks());

  it("appends user message immediately on send", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([{ event: "done", data: "{}" }]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hello"); });

    expect(result.current.messages[0]).toMatchObject({
      role: "user",
      text: "Hello",
    });
  });

  it("accumulates token events into assistant message", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "token", data: '{"text":"Hello"}' },
        { event: "token", data: '{"text":" world"}' },
        { event: "done", data: "{}" },
      ]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });

    await waitFor(() =>
      expect(result.current.messages).toHaveLength(2)
    );

    expect(result.current.messages[1]).toMatchObject({
      role: "assistant",
      text: "Hello world",
    });
  });

  it("inserts blocked message and clears provisional text on blocked event", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "token", data: '{"text":"Some text"}' },
        { event: "blocked", data: '{"reason":"output_blocklist"}' },
      ]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });

    await waitFor(() =>
      expect(result.current.messages).toHaveLength(2)
    );

    expect(result.current.messages[1]).toMatchObject({ role: "blocked" });
    expect(result.current.streaming).toBe(false);
    expect(result.current.provisionalText).toBe("");
  });

  it("sets streaming=false after done", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([{ event: "done", data: "{}" }]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });

    await waitFor(() =>
      expect(result.current.streaming).toBe(false)
    );
  });

  it("inserts error message when stream closes without a terminal event", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "token", data: '{"text":"partial"}' },
        // stream closes here — no done/blocked/error
      ]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });

    await waitFor(() =>
      expect(result.current.streaming).toBe(false)
    );

    expect(result.current.messages[1]).toMatchObject({
      role: "error",
      text: "Connection lost",
    });
    expect(result.current.provisionalText).toBe("");
  });

  it("inserts error message on error event", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "error", data: '{"message":"internal server error"}' },
      ]),
    });

    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });

    await waitFor(() =>
      expect(result.current.messages).toHaveLength(2)
    );

    expect(result.current.messages[1]).toMatchObject({
      role: "error",
      text: "internal server error",
    });
  });
});
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd web && npm test -- useChat
```

- [ ] **Step 3: Implement useChat**

`web/app/chat/[sessionId]/useChat.ts`:
```typescript
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

      const userMsg: Message = {
        id: crypto.randomUUID(),
        role: "user",
        text,
      };
      setMessages((prev) => [...prev, userMsg]);
      setStreaming(true);
      setProvisionalText("");

      let accumulated = "";
      let settled = false; // set to true when done/blocked/error event received

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
          const events = buffer.split("\n\n");
          buffer = events.pop() ?? "";

          for (const raw of events) {
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

        // Stream closed without a terminal event (network drop, server crash)
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
```

- [ ] **Step 4: Run test — expect PASS**

```bash
cd web && npm test -- useChat
```

- [ ] **Step 5: Commit**

```bash
git add app/chat/\[sessionId\]/useChat.ts __tests__/useChat.test.ts
git commit -m "Add useChat hook with SSE streaming and blocked/error handling"
```

---

## Task 6: Scenario picker page

**Files:**
- Create: `web/app/page.tsx`

This is a Server Component: it fetches scenarios on the server, then creates a session via a server action when the user clicks a card.

Each scenario card is wrapped in a `<form>` whose single `<button type="submit">` is the card itself. No separate hidden button, no nested interactive elements.

- [ ] **Step 1: Implement the page**

`web/app/page.tsx`:
```typescript
import { redirect } from "next/navigation";
import { ScenarioCard, type Scenario } from "@/components/ScenarioCard";

async function getScenarios(): Promise<Scenario[]> {
  const apiUrl = process.env.API_URL ?? "http://localhost:8080";
  const res = await fetch(`${apiUrl}/api/scenarios`, { cache: "no-store" });
  if (!res.ok) throw new Error("Failed to fetch scenarios");
  const body = await res.json();
  return body.scenarios as Scenario[];
}

async function createSession(scenarioName: string) {
  "use server";
  const apiUrl = process.env.API_URL ?? "http://localhost:8080";
  const res = await fetch(`${apiUrl}/api/sessions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ scenario_name: scenarioName }),
  });
  if (!res.ok) throw new Error("Failed to create session");
  const body = await res.json();
  redirect(`/chat/${body.session_id}`);
}

export default async function Home() {
  const scenarios = await getScenarios();

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="w-full max-w-xl">
        <h1 className="text-2xl font-bold mb-2">Guardrailed Chatbot</h1>
        <p className="text-muted-foreground mb-8">
          Choose a scenario to start a conversation.
        </p>
        <div className="flex flex-col gap-4">
          {scenarios.map((s) => (
            <form key={s.name} action={createSession.bind(null, s.name)}>
              <button
                type="submit"
                aria-label={s.persona_name}
                className="w-full text-left rounded-xl focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
              >
                <ScenarioCard scenario={s} />
              </button>
            </form>
          ))}
        </div>
      </div>
    </main>
  );
}
```

Each card is a single `<button type="submit">` — one interactive element, one action. Works without JavaScript.

- [ ] **Step 2: Commit**

```bash
git add app/page.tsx
git commit -m "Add scenario picker page"
```

---

## Task 7: Chat page

**Files:**
- Create: `web/app/chat/[sessionId]/page.tsx`

- [ ] **Step 1: Implement the chat page**

`web/app/chat/[sessionId]/page.tsx`:
```typescript
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
```

- [ ] **Step 2: Commit**

```bash
git add app/chat/\[sessionId\]/page.tsx
git commit -m "Add chat page with streaming UI"
```

---

## Task 8: Update README with session limitation

**Files:**
- Modify: `README.md` (at repo root)

- [ ] **Step 1: Add the limitations section**

Find the "## Deployment security recommendations" section in `README.md` and add the following section immediately before it:

```markdown
## Known limitations

**Session URL sharing.** Session IDs are embedded in the chat URL
(`/chat/<session-id>`). Anyone with the link can send messages in your
session and read its history. There is no authentication gate. This is
acceptable for a demo testbed; a production deployment would require
an httpOnly session cookie or equivalent auth mechanism.
```

- [ ] **Step 2: Commit**

```bash
cd ..  # repo root
git add README.md
git commit -m "Document session URL sharing limitation in README"
```

---

## Task 9: Typecheck, lint, and final verification

- [ ] **Step 1: Run all tests**

```bash
cd web && npm test
```

Expected: all tests pass.

- [ ] **Step 2: Typecheck**

```bash
cd web && npx tsc --noEmit
```

Expected: no type errors.

- [ ] **Step 3: Lint**

```bash
cd web && npm run lint
```

Expected: no lint errors.

- [ ] **Step 4: Build**

```bash
cd web && npm run build
```

Expected: build succeeds.

- [ ] **Step 5: Smoke test (requires Go server running)**

Start the Go server, then:

```bash
cd web && npm run dev
```

Open `http://localhost:3000`. Verify:
- Scenario cards load
- Clicking a scenario navigates to `/chat/<uuid>`
- Sending a message streams tokens in real time
- A message that hits the blocklist shows the blocked notice
- Input is disabled while streaming

---

## Security note

The Go server must generate session IDs using `crypto/rand` (UUID v4 or equivalent). The URL-visible session ID is the only session identifier — it cannot be guessed, but it can be shared. See the README limitations section.
