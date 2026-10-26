import Link from "next/link";
import { useRouter } from "next/router";
import { cn } from "@/lib/utils";

const pages = [
  { href: "/admin/reviews", label: "Reviews" },
  { href: "/admin/reports", label: "Reports" },
];

// AdminNav switches between the pages where admins moderate the marketplace.
export default function AdminNav() {
  const router = useRouter();

  return (
    <nav className="flex gap-4 text-sm mb-6">
      {pages.map((p) => (
        <Link
          key={p.href}
          href={p.href}
          className={cn(
            "hover:text-foreground",
            router.pathname === p.href
              ? "text-foreground font-medium"
              : "text-muted-foreground"
          )}
        >
          {p.label}
        </Link>
      ))}
    </nav>
  );
}
