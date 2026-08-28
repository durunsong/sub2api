import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia } from "pinia";
import { createI18n } from "vue-i18n";
import type { SubscriptionPlan } from "@/types/payment";

const labels: Record<string, string> = {
  "payment.days": "天",
  "payment.weeks": "周",
  "payment.months": "月",
  "payment.perMonth": "月",
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

const mountPlanCard = (groupPlatform: string, overrides: Partial<SubscriptionPlan> = {}) =>
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
        ...overrides,
      },
    },
    global: { plugins: [i18n, createPinia()] },
  });

describe("SubscriptionPlanCard", () => {
  it("shows Claude Max for Anthropic plans", () => {
    const text = mountPlanCard("anthropic").text();

    expect(text).toContain("Claude Max");
    expect(text).not.toContain("Anthropic");
  });

  it("uses the same Claude Max badge for Kiro plans", () => {
    const text = mountPlanCard("kiro").text();

    expect(text).toContain("Claude Max");
    expect(text).not.toContain("GLM");
  });

  it("shows GLM for OpenAI-compatible GLM plans while keeping regular OpenAI labels", () => {
    const glmText = mountPlanCard("openai", {
      name: "智普-GLM-月卡-1亿-token",
      description: "支持 GLM Coding",
    }).text();
    const openAIText = mountPlanCard("openai", {
      name: "Codex Pro",
      description: "支持 GPT 模型",
    }).text();

    expect(glmText).toContain("GLM");
    expect(glmText).not.toContain("OpenAI");
    expect(openAIText).toContain("OpenAI");
  });

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

    expect(text).toContain("/ 月");
    expect(text).not.toContain("/ 1天");
  });

  it("uses the configured currency symbol while preserving USD for legacy plans", () => {
    const cnyPlan = mountPlanCard("openai", { currency: "CNY", original_price: 20 }).text();

    expect(cnyPlan).toContain("¥10CNY");
    expect(cnyPlan.replace(/\s/g, "")).toContain("¥20CNY");
    expect(mountPlanCard("openai", { currency: "USD" }).text()).toContain("$10USD");
    expect(mountPlanCard("openai", { currency: "" }).text()).toContain("$10");
  });

  it.each([
    ["long Chinese", "企业全球加速专业订阅套餐（含高级模型与优先支持）"],
    ["long English", "Enterprise Global Acceleration Subscription with Priority Support"],
    ["unbroken token", "EnterpriseGlobalAccelerationSubscriptionWithPrioritySupport1234567890"],
  ])("keeps the full %s plan title accessible in a bounded two-line area", (_label, name) => {
    const wrapper = mountPlanCard("openai", { name });
    const title = wrapper.get("h3");

    expect(title.text()).toBe(name);
    expect(title.attributes("title")).toBe(name);
    expect(title.classes()).toEqual(expect.arrayContaining([
      "min-w-0",
      "flex-1",
      "break-words",
      "line-clamp-2",
      "[overflow-wrap:anywhere]",
      "text-xl",
    ]));
    expect(title.classes()).not.toContain("truncate");
  });

  it("keeps title, badge, price, description, and purchase action readable", () => {
    const wrapper = mountPlanCard("openai", {
      name: "Enterprise Global Acceleration Subscription with Priority Support",
      price: 123.45,
      currency: "USD",
      description: "Includes advanced models and priority support.",
    });
    const title = wrapper.get("h3");
    const badge = wrapper.findAll("span").find((node) => node.text() === "OpenAI");
    const price = wrapper.findAll("span").find((node) => node.text() === "123.45");

    expect(title.classes()).toContain("min-w-0");
    expect(title.classes()).toContain("flex-1");
    expect(badge?.classes()).toContain("shrink-0");
    expect(title.element.parentElement?.classList).toContain("min-w-0");
    expect(price?.text()).toBe("123.45");
    expect(wrapper.get("p").text()).toBe("Includes advanced models and priority support.");
    expect(wrapper.get("button").text().toLowerCase()).toMatch(/subscribe|payment\.subscribenow/);
  });

  it("keeps short plan titles compact and accessible via native title", () => {
    const wrapper = mountPlanCard("openai", { name: "Pro", description: "" });
    const title = wrapper.get("h3");

    expect(title.text()).toBe("Pro");
    expect(title.attributes("title")).toBe("Pro");
    expect(title.classes()).toEqual(expect.arrayContaining(["text-xl", "font-bold", "line-clamp-2"]));
  });
});
