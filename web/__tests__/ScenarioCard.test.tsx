import { render, screen } from "@testing-library/react";
import { ScenarioCard, type Scenario } from "@/components/ScenarioCard";

const scenario: Scenario = {
  name: "financial_advisor",
  persona_name: "Morgan",
  persona_description: "A cautious financial literacy assistant.",
};

describe("ScenarioCard", () => {
  it("renders persona name and description", () => {
    render(<ScenarioCard scenario={scenario} />);
    expect(screen.getByText("Morgan")).toBeInTheDocument();
    expect(screen.getByText(/cautious financial/)).toBeInTheDocument();
  });
});
