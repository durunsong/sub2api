package service

// Fork：未活跃用户余额券邮件注册为可退订的通知邮件事件，模板在 /admin/settings 邮件模板中维护。
// 独立文件 + init 注册，避免改动上游事件表。

const NotificationEmailEventRedeemInactiveCoupon = "redeem.inactive_coupon"

var redeemEmailCouponPlaceholders = []string{"coupon_value", "redeem_code", "expires_at", "unsubscribe_url"}

func init() {
	notificationEmailEventOrder = append(notificationEmailEventOrder, NotificationEmailEventRedeemInactiveCoupon)
	notificationEmailEventDefinitions[NotificationEmailEventRedeemInactiveCoupon] = NotificationEmailEventInfo{
		Event:        NotificationEmailEventRedeemInactiveCoupon,
		Label:        "Inactive user balance coupon",
		Description:  "Sent by the admin redeem-code campaign to inactive users with a personal balance redeem code.",
		Category:     "marketing",
		Optional:     true,
		Placeholders: append(append([]string{}, notificationEmailCommonPlaceholders...), redeemEmailCouponPlaceholders...),
	}
	notificationEmailOfficialTemplates[NotificationEmailEventRedeemInactiveCoupon] = map[string]notificationEmailOfficialTemplate{
		notificationEmailDefaultLocale: {
			Subject: "[{{site_name}}] A {{coupon_value}} balance coupon for you",
			HTML: notificationEmailCard("#0878ed", "We miss you", `
<p>Hello {{recipient_name}},</p>
<p>It has been a while since your last visit. Here is a <strong>{{coupon_value}}</strong> balance coupon to welcome you back.</p>
<div class="code-box"><p style="margin:0;color:#0878ed;font-size:20px;line-height:1.4;font-weight:800;letter-spacing:2px;word-break:break-all;font-family:Menlo,Consolas,monospace;">{{redeem_code}}</p></div>
<p>Sign in to {{site_name}} and enter the code on the Redeem page to add the balance. Valid until: <strong>{{expires_at}}</strong>.</p>
<p>This code is personal. Please do not share it.</p>
<p class="muted"><a href="{{unsubscribe_url}}">Unsubscribe from coupon emails</a></p>`),
		},
		notificationEmailLocaleChinese: {
			Subject: "[{{site_name}}] 送您一张 {{coupon_value}} 余额券",
			HTML: notificationEmailCard("#0878ed", "好久不见", `
<p>{{recipient_name}}，您好：</p>
<p>好久没见到您了，我们为您准备了一张 <strong>{{coupon_value}}</strong> 的余额兑换券，欢迎回来继续使用。</p>
<div class="code-box"><p style="margin:0;color:#0878ed;font-size:20px;line-height:1.4;font-weight:800;letter-spacing:2px;word-break:break-all;font-family:Menlo,Consolas,monospace;">{{redeem_code}}</p></div>
<p>登录 {{site_name}} 后，在「兑换」页面输入兑换码即可到账。有效期至：<strong>{{expires_at}}</strong>。</p>
<p>兑换码仅限本人使用，请勿转发他人。</p>
<p class="muted"><a href="{{unsubscribe_url}}">退订此类优惠券邮件</a></p>`),
		},
	}
}

// notificationEmailForkSampleVariables 为 Fork 事件补充预览示例值。
func notificationEmailForkSampleVariables(event string) map[string]string {
	if event != NotificationEmailEventRedeemInactiveCoupon {
		return nil
	}
	return map[string]string{
		"coupon_value": "$5.00",
		"redeem_code":  "PREVIEW-CODE-0000",
		"expires_at":   "2026-12-31 23:59 UTC",
	}
}

func redeemEmailNeverExpiresLabel(locale string) string {
	if normalizeNotificationLocale(locale) == notificationEmailLocaleChinese {
		return "长期有效"
	}
	return "Never expires"
}
