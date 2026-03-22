import { renderHook, act, waitFor } from "@testing-library/react";
import { useChat } from "@/app/chat/[sessionId]/useChat";

// Server emits: data: {"type":"...","content":"..."}\n\n
// No named SSE event: lines — type is in the JSON payload.
function makeSSEStream(events: Array<{ type: string; content?: string }>) {
  const encoder = new TextEncoder();
  const chunks = events.map(({ type, content }) => {
    const payload = content !== undefined
      ? JSON.stringify({ type, content })
      : JSON.stringify({ type });
    return encoder.encode(`data: ${payload}\n\n`);
  });
  let i = 0;
  return new ReadableStream({
    pull(controller) {
      if (i < chunks.length) controller.enqueue(chunks[i++]);
      else controller.close();
    },
  });
}

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("useChat", () => {
  beforeEach(() => fetchMock.mockReset());

  it("appends user message immediately on send", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([{ type: "done" }]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hello"); });
    expect(result.current.messages[0]).toMatchObject({ role: "user", text: "Hello" });
  });

  it("accumulates token events into assistant message on done", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { type: "token", content: "Hello" },
        { type: "token", content: " world" },
        { type: "done" },
      ]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.messages).toHaveLength(2));
    expect(result.current.messages[1]).toMatchObject({ role: "assistant", text: "Hello world" });
  });

  it("inserts blocked message and resets on blocked event", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { type: "token", content: "Some text" },
        { type: "blocked" },
      ]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.messages).toHaveLength(2));
    expect(result.current.messages[1]).toMatchObject({ role: "blocked" });
    expect(result.current.streaming).toBe(false);
    expect(result.current.provisionalText).toBe("");
  });

  it("sets streaming=false after done", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([{ type: "done" }]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.streaming).toBe(false));
  });

  it("inserts error when stream closes without terminal event", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { type: "token", content: "partial" },
        // stream closes without done or blocked
      ]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.streaming).toBe(false));
    expect(result.current.messages[1]).toMatchObject({ role: "error", text: "Connection lost" });
    expect(result.current.provisionalText).toBe("");
  });

  it("inserts error message on HTTP error", async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 500, body: null });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.messages).toHaveLength(2));
    expect(result.current.messages[1]).toMatchObject({ role: "error" });
    expect(result.current.streaming).toBe(false);
  });
});
