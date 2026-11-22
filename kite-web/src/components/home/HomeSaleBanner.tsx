import { saleLastDay } from "@/lib/billing";
import { useBillingSale } from "@/lib/hooks/api";
import Link from "next/link";

// HomeSaleBanner is a strip above the navbar while a sale runs.
export default function HomeSaleBanner() {
  const sale = useBillingSale();
  if (!sale) return null;

  return (
    <Link
      href="/pricing"
      className="block bg-primary px-4 py-2 text-center text-sm font-medium text-primary-foreground hover:bg-primary/90"
    >
      {sale.name}: {sale.percent}% off the first year of yearly plans until{" "}
      {saleLastDay(sale)}, with the code {sale.code}.{" "}
      <span className="underline">See the plans</span>
    </Link>
  );
}
