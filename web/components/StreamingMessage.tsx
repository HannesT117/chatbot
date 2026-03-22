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
