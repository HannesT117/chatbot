import { Alert, AlertDescription } from "@/components/ui/alert";
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
