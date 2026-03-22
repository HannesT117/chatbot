"use client";

export default function Error({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  return (
    <main className="flex min-h-screen items-center justify-center p-8">
      <div className="text-center">
        <h1 className="text-xl font-semibold mb-2">Something went wrong</h1>
        <p className="text-muted-foreground text-sm">{error.message}</p>
        <p className="text-muted-foreground text-xs mt-1">Is the Go server running?</p>
        <button
          onClick={() => reset()}
          className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm"
        >
          Try again
        </button>
      </div>
    </main>
  );
}
