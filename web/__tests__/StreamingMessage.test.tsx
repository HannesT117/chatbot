import { render } from "@testing-library/react";
import { StreamingMessage } from "@/components/StreamingMessage";

describe("StreamingMessage", () => {
  it("renders partial text with a cursor", () => {
    render(<StreamingMessage text="Hello wor" />);
    expect(document.querySelector("[aria-label='typing cursor']")).toBeInTheDocument();
  });

  it("renders empty string with just the cursor", () => {
    render(<StreamingMessage text="" />);
    expect(document.querySelector("[aria-label='typing cursor']")).toBeInTheDocument();
  });
});
