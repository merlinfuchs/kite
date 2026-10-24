import { type ClassValue, clsx } from "clsx";
import { formatDuration, formatRelative, intervalToDuration } from "date-fns";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function abbreviateName(name: string): string {
  const parts = name.split(" ");
  return parts
    .map((part) => part.charAt(0).toUpperCase())
    .join("")
    .slice(0, 2);
}

export function readFileAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const dataURL = reader.result as string;
      const data = dataURL.split(",")[1];
      resolve(data);
    };
    reader.onerror = (error) => {
      reject(error);
    };
    reader.readAsDataURL(file);
  });
}

export function getUniqueId(): number {
  return Math.floor(Math.random() * 1000000000);
}

export function formatTime(date: Date): string {
  return date.toLocaleTimeString();
}

export function formatDate(date: Date): string {
  return date.toLocaleDateString();
}

export function formatDateTime(date: Date): string {
  return formatRelative(date, new Date());
}

export function formatNumber(x: number | undefined | null) {
  if (!x) return "0";
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

export function formatInterval(seconds: number) {
  return formatDuration(intervalToDuration({ start: 0, end: seconds * 1000 }));
}

// Credits and AI prompts are counted per calendar month in UTC.
export function monthlyResetDate(now = new Date()): Date {
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1));
}

export function formatMonthlyResetDate(): string {
  return formatUTCDay(monthlyResetDate());
}

// formatUTCDay formats a date like "November 1" on the UTC calendar, which
// credits are counted by.
export function formatUTCDay(date: Date): string {
  return date.toLocaleDateString(undefined, {
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });
}
