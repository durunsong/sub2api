import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia } from "pinia";
import { createI18n } from "vue-i18n";

const labels: Record<string, string> = {
  "payment.admin.days": "天",
  "payment.admin.weeks": "周",
  "payment.admin.months": "月",
  "payment.planCard.quota": "Quota",
  "payment.planCard.rate": "Rate",
  "payment.planCard.unlimited": "Unlimited",
  "payment.planCard.dailyLimit": "Daily",
  "payment.planCard.weeklyLimit": "Weekly",
  "payment.planCard.monthlyLimit": "Monthly",
  "payment.subscribeNow": "Subscribe now",
  "payment.renewNow": "Renew",
};

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => labels[key] ?? key,
    }),
  };
});

import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const i18n = createI18n({
  legacy: false,
  locale: "zh",
  fallbackWarn: false,
  missingWarn: false,
  messages: { zh: labels },
});

const mountPlanCard = (
  groupPlatform: string,
  overrides: Partial<{
    validity_days: number;
    validity_unit: string;
    name: string;
  }> = {},
) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        id: 1,
        group_id: 10,
        group_platform: groupPlatform,
        name: overrides.name ?? "Pro",
        price: 10,
        amount: 1000,
        features: [],
        rate_multiplier: 1,
        validity_days: overrides.validity_days ?? 30,
        validity_unit: overrides.validity_unit ?? "day",
        supported_model_scopes: ["claude", "gemini_text", "gemini_image"],
        is_active: true,
      },
    },
    global: { plugins: [i18n, createPinia()] },
  });

describe("SubscriptionPlanCard", () => {
  it("does not show Antigravity model scopes for OpenAI plans", () => {
    const text = mountPlanCard("openai").text();

    expect(text).not.toContain("Claude");
    expect(text).not.toContain("Gemini");
    expect(text).not.toContain("Imagen");
  });

  it("shows model scopes for Antigravity plans", () => {
    const text = mountPlanCard("antigravity").text();

    expect(text).toContain("Claude");
    expect(text).toContain("Gemini");
    expect(text).toContain("Imagen");
  });

  it("shows week validity suffix for weekly plans", () => {
    const text = mountPlanCard("openai", {
      validity_days: 1,
      validity_unit: "weeks",
    }).text();

    expect(text).toContain("/ 1周");
    expect(text).not.toContain("/ 1天");
  });

  it("shows month validity suffix for monthly plans", () => {
    const text = mountPlanCard("openai", {
      validity_days: 1,
      validity_unit: "months",
    }).text();

    expect(text).toContain("/ 1月");
    expect(text).not.toContain("/ 1天");
  });
});
