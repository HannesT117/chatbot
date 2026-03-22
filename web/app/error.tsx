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
