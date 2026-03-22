import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ChatInput } from "@/components/ChatInput";
import { describe, it, expect, vi } from "vitest";

describe("ChatInput", () => {
  it("calls onSend with the message text and clears the input", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.type(screen.getByRole("textbox"), "Hello there");
    await user.click(screen.getByRole("button", { name: /send/i }));
    expect(onSend).toHaveBeenCalledWith("Hello there");
    expect(screen.getByRole("textbox")).toHaveValue("");
  });

  it("submits on Enter (without shift)", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.type(screen.getByRole("textbox"), "Hello{Enter}");
    expect(onSend).toHaveBeenCalledWith("Hello");
  });

  it("does not submit on Shift+Enter", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();
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
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} disabled={false} />);
    await user.click(screen.getByRole("button", { name: /send/i }));
    expect(onSend).not.toHaveBeenCalled();
  });
});
