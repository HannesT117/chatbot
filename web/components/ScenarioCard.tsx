import { Card, CardHeader, CardDescription } from "@/components/ui/card";

export type Scenario = {
  id: string;
  name: string;
  persona_name: string;
};

type Props = {
  scenario: Scenario;
};

// ScenarioCard is rendered inside a <button>. Do NOT use heading elements
// (CardTitle renders <h3>) — headings inside buttons are invalid HTML.
export function ScenarioCard({ scenario }: Props) {
  return (
    <Card className="w-full hover:border-primary transition-colors">
      <CardHeader>
        <p className="font-semibold text-base">{scenario.persona_name}</p>
        <CardDescription>{scenario.name}</CardDescription>
      </CardHeader>
    </Card>
  );
}
