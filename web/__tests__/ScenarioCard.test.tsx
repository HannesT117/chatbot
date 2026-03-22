import { render, screen } from "@testing-library/react";
import { ScenarioCard, type Scenario } from "@/components/ScenarioCard";

const scenario: Scenario = {
  id: "financial_advisor",
  name: "Financial Advisor",
  persona_name: "Morgan",
};

describe("ScenarioCard", () => {
  it("renders persona name and scenario name", () => {
    render(<ScenarioCard scenario={scenario} />);
    expect(screen.getByText("Morgan")).toBeInTheDocument();
    expect(screen.getByText("Financial Advisor")).toBeInTheDocument();
  });
});
