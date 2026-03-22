import { render, screen } from "@testing-library/react";
import { ChatMessage, type Message } from "@/components/ChatMessage";

describe("ChatMessage", () => {
  it("renders user message aligned right", () => {
    const msg: Message = { id: "1", role: "user", text: "Hello" };
    render(<ChatMessage message={msg} />);
    expect(screen.getByText("Hello")).toBeInTheDocument();
    expect(screen.getByText("Hello").closest("[data-role='user']")).toBeTruthy();
  });

  it("renders assistant message", () => {
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
