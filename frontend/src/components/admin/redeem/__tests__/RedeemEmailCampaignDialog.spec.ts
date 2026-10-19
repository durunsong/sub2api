import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CampaignDialog from "../RedeemEmailCampaignDialog.vue";

const api = vi.hoisted(() => ({
  defaults: vi.fn(),
  search: vi.fn(),
  preview: vi.fn(),
  create: vi.fn(),
}));
vi.mock("@/api/admin", () => ({
  adminAPI: {
    redeem: {
      getEmailDefaults: api.defaults,
      searchEmailRecipients: api.search,
      previewEmail: api.preview,
      createEmailBatch: api.create,
    },
  },
}));
vi.mock("vue-i18n", () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
    locale: { value: "zh-CN" },
  }),
}));
vi.mock("@/utils/format", () => ({
  formatDateTime: (value: string) => `formatted:${value}`,
}));

const defaults = {
  template_event: "redeem.inactive_coupon",
  segment_options: [
    "inactive_1_2m",
    "inactive_2_3m",
    "inactive_3m_plus",
    "never_used",
    "zero_balance",
  ],
  default_segments: ["inactive_1_2m", "inactive_2_3m", "inactive_3m_plus"],
  rate_per_second: 1,
  smtp_configured: true,
  max_recipients: 1000,
  default_expires_in_days: 30,
  max_expires_in_days: 3650,
};

const recipients = [
  {
    user_id: 7,
    email: "b@example.com",
    username: "bob",
    last_used_at: "2026-06-01T00:00:00Z",
    segment: "inactive_3m_plus",
    unsubscribed: false,
  },
  {
    user_id: 3,
    email: "a@example.com",
    username: "",
    last_used_at: "2026-08-20T00:00:00Z",
    segment: "inactive_1_2m",
    unsubscribed: false,
  },
  {
    user_id: 9,
    email: "gone@example.com",
    username: "",
    last_used_at: "2026-08-01T00:00:00Z",
    segment: "inactive_2_3m",
    unsubscribed: true,
  },
];

const mountDialog = () =>
  mount(CampaignDialog, {
    props: { show: true },
    global: {
      stubs: {
        BaseDialog: {
          props: ["show", "title"],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>',
        },
      },
    },
  });

async function next(wrapper: ReturnType<typeof mountDialog>) {
  await wrapper.get('[data-test="next-step"]').trigger("click");
  await flushPromises();
}

async function goToRecipients(wrapper: ReturnType<typeof mountDialog>) {
  await flushPromises();
  await next(wrapper);
}

async function searchAndContinue(wrapper: ReturnType<typeof mountDialog>) {
  await wrapper.get('[data-test="search-recipients"]').trigger("click");
  await flushPromises();
  await next(wrapper);
}

describe("RedeemEmailCampaignDialog", () => {
  beforeEach(() => {
    api.defaults.mockReset().mockResolvedValue(structuredClone(defaults));
    api.search
      .mockReset()
      .mockResolvedValue({
        items: structuredClone(recipients),
        total: 3,
        max_recipients: 1000,
      });
    api.preview
      .mockReset()
      .mockResolvedValue({
        subject: "Hi Demo",
        html: "<p>PREVIEW</p><script>alert(1)</script>",
      });
    api.create
      .mockReset()
      .mockResolvedValue({ id: 12, status: "pending", total_count: 2 });
  });

  it("walks coupon, segments, settings-template preview and confirm steps, then submits an idempotent batch", async () => {
    const wrapper = mountDialog();
    await goToRecipients(wrapper);

    expect(
      wrapper
        .get('[data-test="segment-inactive_1_2m"]')
        .attributes("aria-pressed"),
    ).toBe("true");
    expect(
      wrapper
        .get('[data-test="segment-never_used"]')
        .attributes("aria-pressed"),
    ).toBe("false");

    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    expect(api.search).toHaveBeenCalledWith({
      segments: ["inactive_1_2m", "inactive_2_3m", "inactive_3m_plus"],
      search: undefined,
    });
    expect(wrapper.text()).toContain(
      "admin.redeemEmail.recipients.segment.inactive_3m_plus",
    );
    expect(wrapper.text()).toContain("formatted:2026-06-01T00:00:00Z");
    expect(wrapper.find('[data-test="unsubscribed-9"]').exists()).toBe(true);
    expect(
      (wrapper.get('[data-test="recipient-9"]').element as HTMLInputElement)
        .checked,
    ).toBe(false);
    expect(wrapper.get('[data-test="selected-count"]').text()).toContain(
      '"selected":2',
    );

    await next(wrapper);
    expect(api.preview).toHaveBeenLastCalledWith({
      coupon_value: 5,
      code_expires_in_days: 30,
      locale: "zh",
    });
    expect(
      wrapper.get('[data-test="edit-template-link"]').attributes("href"),
    ).toBe("/admin/settings?tab=email&email_event=redeem.inactive_coupon");
    const frame = wrapper.get('[data-test="email-preview"] iframe');
    expect(frame.attributes("sandbox")).toBe("");
    expect(frame.attributes("srcdoc")).toContain("PREVIEW");

    await wrapper.get('[data-test="preview-locale-en"]').trigger("click");
    await flushPromises();
    expect(api.preview).toHaveBeenLastCalledWith({
      coupon_value: 5,
      code_expires_in_days: 30,
      locale: "en",
    });

    await next(wrapper);
    expect(wrapper.get('[data-test="confirm-count"]').text()).toBe("2");
    expect(wrapper.get('[data-test="confirm-total"]').text()).toBe("$10.00");
    expect(wrapper.get('[data-test="confirm-segments"]').text()).toBe(
      "admin.redeemEmail.recipients.segment.inactive_1_2m、admin.redeemEmail.recipients.segment.inactive_2_3m、admin.redeemEmail.recipients.segment.inactive_3m_plus",
    );
    expect(wrapper.find('[data-test="confirm-unsubscribed"]').exists()).toBe(
      false,
    );

    api.create.mockRejectedValueOnce({ message: "network down" });
    await wrapper.get('[data-test="submit-campaign"]').trigger("click");
    await flushPromises();
    await wrapper.get('[data-test="submit-campaign"]').trigger("click");
    await flushPromises();

    const [firstPayload, firstKey] = api.create.mock.calls[0];
    const [, secondKey] = api.create.mock.calls[1];
    expect(firstPayload).toEqual({
      coupon_value: 5,
      code_expires_in_days: 30,
      segments: ["inactive_1_2m", "inactive_2_3m", "inactive_3m_plus"],
      user_ids: [3, 7],
    });
    expect(firstKey).toBeTruthy();
    expect(secondKey).toBe(firstKey);
    expect(wrapper.emitted("created")?.[0]).toEqual([
      { id: 12, status: "pending", total_count: 2 },
    ]);
  });

  it("toggles segments as a multi-select, keeps at least one, and clears stale results", async () => {
    const wrapper = mountDialog();
    await goToRecipients(wrapper);
    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    expect(wrapper.find('[data-test="selected-count"]').exists()).toBe(true);

    await wrapper.get('[data-test="segment-never_used"]').trigger("click");
    await wrapper.get('[data-test="segment-inactive_1_2m"]').trigger("click");
    await flushPromises();
    expect(wrapper.find('[data-test="selected-count"]').exists()).toBe(false);

    for (const s of ["inactive_2_3m", "inactive_3m_plus", "never_used"]) {
      await wrapper.get(`[data-test="segment-${s}"]`).trigger("click");
    }
    expect(
      wrapper
        .get('[data-test="segment-never_used"]')
        .attributes("aria-pressed"),
    ).toBe("true");

    api.search.mockRejectedValueOnce({
      reason: "REDEEM_EMAIL_RECIPIENT_LIMIT_EXCEEDED",
      message: "too many",
    });
    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    expect(api.search).toHaveBeenLastCalledWith({
      segments: ["never_used"],
      search: undefined,
    });
    expect(wrapper.get('[data-test="campaign-error"]').text()).toContain(
      "admin.redeemEmail.recipients.limitExceeded",
    );
  });

  it("selects zero balance alone and drops inactivity segments", async () => {
    const wrapper = mountDialog();
    await goToRecipients(wrapper);
    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();

    await wrapper.get('[data-test="segment-zero_balance"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-test="segment-zero_balance"]').attributes("aria-pressed")).toBe("true");
    expect(wrapper.get('[data-test="segment-inactive_1_2m"]').attributes("aria-pressed")).toBe("false");
    expect(wrapper.find('[data-test="selected-count"]').exists()).toBe(false);

    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    expect(api.search).toHaveBeenLastCalledWith({
      segments: ["zero_balance"],
      search: undefined,
    });

    await wrapper.get('[data-test="segment-inactive_2_3m"]').trigger("click");
    expect(wrapper.get('[data-test="segment-zero_balance"]').attributes("aria-pressed")).toBe("false");
    expect(wrapper.get('[data-test="segment-inactive_2_3m"]').attributes("aria-pressed")).toBe("true");
    expect(wrapper.get('[data-test="segment-inactive_1_2m"]').attributes("aria-pressed")).toBe("false");
  });

  it("sends never_expires instead of days, warns about unsubscribed picks and uses a new key when the payload changes", async () => {
    const wrapper = mountDialog();
    await flushPromises();
    await wrapper.get('[data-test="never-expires"]').setValue(true);
    await next(wrapper);
    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    await wrapper.get('[data-test="select-all"]').trigger("click");
    await next(wrapper);
    expect(api.preview).toHaveBeenLastCalledWith({
      coupon_value: 5,
      never_expires: true,
      locale: "zh",
    });
    await next(wrapper);
    expect(wrapper.find('[data-test="confirm-unsubscribed"]').exists()).toBe(
      true,
    );

    api.create.mockRejectedValueOnce({
      reason: "REDEEM_EMAIL_RECIPIENT_INELIGIBLE",
      message: "x",
    });
    await wrapper.get('[data-test="submit-campaign"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-test="campaign-error"]').text()).toContain(
      "admin.redeemEmail.errors.REDEEM_EMAIL_RECIPIENT_INELIGIBLE",
    );
    expect(api.create.mock.calls[0][0]).toMatchObject({
      never_expires: true,
      user_ids: [3, 7, 9],
    });
    expect(api.create.mock.calls[0][0]).not.toHaveProperty(
      "code_expires_in_days",
    );

    await wrapper.get('[data-test="prev-step"]').trigger("click");
    await wrapper.get('[data-test="prev-step"]').trigger("click");
    await wrapper.get('[data-test="recipient-9"]').trigger("change");
    await next(wrapper);
    await next(wrapper);
    await wrapper.get('[data-test="submit-campaign"]').trigger("click");
    await flushPromises();
    expect(api.create.mock.calls[1][0].user_ids).toEqual([3, 7]);
    expect(api.create.mock.calls[1][1]).not.toBe(api.create.mock.calls[0][1]);
  });

  it("blocks invalid coupon values, empty selections and a broken settings template", async () => {
    const wrapper = mountDialog();
    await flushPromises();
    await wrapper.get('[data-test="coupon-value"]').setValue("0.123456789");
    await wrapper.get('[data-test="next-step"]').trigger("click");
    expect(wrapper.get('[data-test="campaign-error"]').text()).toBe(
      "admin.redeemEmail.coupon.valueInvalid",
    );

    await wrapper.get('[data-test="coupon-value"]').setValue("2.5");
    await next(wrapper);
    await wrapper.get('[data-test="search-recipients"]').trigger("click");
    await flushPromises();
    await wrapper.get('[data-test="deselect-all"]').trigger("click");
    await wrapper.get('[data-test="next-step"]').trigger("click");
    expect(wrapper.get('[data-test="campaign-error"]').text()).toBe(
      "admin.redeemEmail.recipients.selectAtLeastOne",
    );

    await wrapper.get('[data-test="select-all"]').trigger("click");
    api.preview.mockRejectedValueOnce({
      reason: "REDEEM_EMAIL_TEMPLATE_INVALID",
      message: "bad",
    });
    await next(wrapper);
    expect(wrapper.get('[data-test="campaign-error"]').text()).toContain(
      "admin.redeemEmail.errors.REDEEM_EMAIL_TEMPLATE_INVALID",
    );
    expect(wrapper.find('[data-test="search-recipients"]').exists()).toBe(true);
  });

  it("warns and disables sending when SMTP is not configured", async () => {
    api.defaults.mockResolvedValue({
      ...structuredClone(defaults),
      smtp_configured: false,
    });
    const wrapper = mountDialog();
    await goToRecipients(wrapper);
    await searchAndContinue(wrapper);
    await next(wrapper);

    expect(wrapper.find('[data-test="smtp-warning"]').exists()).toBe(true);
    expect(
      wrapper.get('[data-test="submit-campaign"]').attributes("disabled"),
    ).toBeDefined();
  });
});
