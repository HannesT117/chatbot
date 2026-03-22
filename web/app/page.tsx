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
