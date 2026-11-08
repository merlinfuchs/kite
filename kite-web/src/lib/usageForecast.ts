import { monthlyResetDate } from "./utils";

export const dayMs = 24 * 60 * 60 * 1000;

// The rate is the average of the last full days, so one busy day early in the
// month doesn't project nonsense.
const rateDays = 7;
const minRateDays = 3;

export interface UsageForecast {
  dailyRate: number;
  // The credits used by the end of the month at the current rate.
  projected: number;
  // When the credits run out at the current rate, if before the month ends.
  runsOutAt?: Date;
}

// forecastUsage projects the credits of this month from the credits by day,
// which are UTC days like the monthly reset. Returns null without enough full
// days or once the credits ran out.
export function forecastUsage(
  days: { date: string; credits_used: number }[],
  creditsUsed: number,
  limit: number,
  now = new Date()
): UsageForecast | null {
  if (limit > 0 && creditsUsed >= limit) return null;

  const today = Date.UTC(
    now.getUTCFullYear(),
    now.getUTCMonth(),
    now.getUTCDate()
  );
  // Days before the app first used credits this month, like before it was
  // created, would pull the rate down.
  const firstUsed = days.findIndex((d) => d.credits_used > 0);
  const fullDays = days
    .slice(firstUsed < 0 ? days.length : firstUsed)
    .filter((d) => new Date(d.date).getTime() < today)
    .slice(-rateDays);
  if (fullDays.length < minRateDays) return null;

  const dailyRate =
    fullDays.reduce((sum, d) => sum + d.credits_used, 0) / fullDays.length;

  const daysLeft = (monthlyResetDate(now).getTime() - now.getTime()) / dayMs;
  const projected = Math.round(creditsUsed + dailyRate * daysLeft);

  let runsOutAt: Date | undefined;
  if (limit > 0 && projected >= limit && dailyRate > 0) {
    runsOutAt = new Date(
      now.getTime() + ((limit - creditsUsed) / dailyRate) * dayMs
    );
  }

  return { dailyRate, projected, runsOutAt };
}
