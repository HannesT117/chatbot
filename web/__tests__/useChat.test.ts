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

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("useChat", () => {
  beforeEach(() => fetchMock.mockReset());

  it("appends user message immediately on send", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([{ event: "done", data: "{}" }]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hello"); });
    expect(result.current.messages[0]).toMatchObject({ role: "user", text: "Hello" });
  });

  it("accumulates token events into assistant message on done", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "token", data: '{"text":"Hello"}' },
        { event: "token", data: '{"text":" world"}' },
        { event: "done", data: "{}" },
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
        { event: "token", data: '{"text":"Some text"}' },
        { event: "blocked", data: '{"reason":"output_blocklist"}' },
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
      body: makeSSEStream([{ event: "done", data: "{}" }]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.streaming).toBe(false));
  });

  it("inserts error message on error event", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "error", data: '{"message":"internal server error"}' },
      ]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.messages).toHaveLength(2));
    expect(result.current.messages[1]).toMatchObject({ role: "error", text: "internal server error" });
  });

  it("inserts error when stream closes without terminal event", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      body: makeSSEStream([
        { event: "token", data: '{"text":"partial"}' },
        // stream closes without done/blocked/error
      ]),
    });
    const { result } = renderHook(() => useChat("session-1"));
    act(() => { result.current.send("Hi"); });
    await waitFor(() => expect(result.current.streaming).toBe(false));
    expect(result.current.messages[1]).toMatchObject({ role: "error", text: "Connection lost" });
    expect(result.current.provisionalText).toBe("");
  });
});
