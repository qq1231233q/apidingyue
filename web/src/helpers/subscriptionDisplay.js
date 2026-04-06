export function getSubscriptionSourceLabel(source, t) {
  switch (source) {
    case 'code':
      return t('激活码兑换');
    case 'admin':
      return t('后台发放');
    case 'order':
      return t('购买套餐');
    default:
      return source || t('未知来源');
  }
}

function resolveSubscriptionPlanTitle(subscription, summaryPlan, planTitleMap) {
  const planId = subscription?.plan_id;
  if (!planId) {
    return '';
  }
  return summaryPlan?.title || planTitleMap?.get?.(planId) || '';
}

export function getSubscriptionPlanLabel(
  subscription,
  summaryPlan,
  planTitleMap,
  t,
) {
  const planId = subscription?.plan_id;
  const planTitle = resolveSubscriptionPlanTitle(
    subscription,
    summaryPlan,
    planTitleMap,
  );

  if (planTitle) {
    return planTitle;
  }
  if (planId) {
    return `${t('套餐')} #${planId}`;
  }
  if (subscription?.source === 'code') {
    return t('激活码订阅');
  }
  return t('订阅实例');
}

export function getSubscriptionDisplayTitle(
  subscription,
  summaryPlan,
  planTitleMap,
  t,
) {
  const planId = subscription?.plan_id;
  const planTitle = resolveSubscriptionPlanTitle(
    subscription,
    summaryPlan,
    planTitleMap,
  );
  const subscriptionId = subscription?.id;

  if (planTitle) {
    return `${planTitle} · ${t('订阅')} #${subscriptionId}`;
  }
  if (planId) {
    return `${t('套餐')} #${planId} · ${t('订阅')} #${subscriptionId}`;
  }
  if (subscription?.source === 'code') {
    return `${t('激活码订阅')} #${subscriptionId}`;
  }
  return `${t('订阅')} #${subscriptionId}`;
}

export function getSubscriptionAvailableGroup(subscription, summaryPlan) {
  return subscription?.available_group || summaryPlan?.available_group || '';
}

export function getSubscriptionStatusMeta(subscription, now = Date.now() / 1000) {
  const end = subscription?.end_time || 0;
  const status = subscription?.status || '';
  const isExpired = end > 0 && end < now;
  const isCancelled = status === 'cancelled';
  const isActive = status === 'active' && !isExpired;

  return {
    isActive,
    isCancelled,
    isExpired,
  };
}
