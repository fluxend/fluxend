import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { dbCookie, organizationCookie } from "./cookies";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function isClient() {
  return typeof window !== "undefined";
}

export function isServer() {
  return !isClient();
}

export function getProjectApiUrl(dbId: string): string {
  return `${import.meta.env.VITE_FLX_API_URL}/rest/${dbId}/`;
}

export function getDbIdCookie(headers: Headers): Promise<string> {
  const cookies = headers.get("cookie");
  return dbCookie.parse(cookies);
}

export function getOrganizationIdCookie(headers: Headers): Promise<string> {
  const cookies = headers.get("cookie");
  return organizationCookie.parse(cookies);
}

export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export function bytesToMB(bytes: number): number {
  return bytes / (1024 * 1024);
}

export function mbToBytes(mb: number): number {
  return mb * 1024 * 1024;
}

// TODO: Validate Response with Types using zod
export async function getTypedResponseData<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    // No content response, return empty object
    return { ok: response.ok, status: response.status } as T;
  }
  const json = await response.json();
  return { ok: response.ok, status: response.status, ...json } as T;
}

export function formatTimestamp(timestamp: string): {
  date: string;
  time: string;
  fullDate: string;
  relativeTime: string;
} {
  try {
    // Parse the timestamp - if it doesn't end with 'Z', assume it's UTC and add it
    const utcTimestamp = timestamp.endsWith('Z') ? timestamp : `${timestamp}Z`;
    const date = new Date(utcTimestamp);

    // Check if date is valid
    if (isNaN(date.getTime())) {
      return { date: "Invalid date", time: "", fullDate: "", relativeTime: "" };
    }

    // Now date is in user's local timezone
    // Format date as "Mon DD, YYYY"
    const monthNames = [
      "Jan",
      "Feb",
      "Mar",
      "Apr",
      "May",
      "Jun",
      "Jul",
      "Aug",
      "Sep",
      "Oct",
      "Nov",
      "Dec",
    ];
    const month = monthNames[date.getMonth()];
    const day = date.getDate();
    const year = date.getFullYear();

    // Format time as "HH:MM AM/PM" in local timezone
    let hours = date.getHours();
    const minutes = date.getMinutes().toString().padStart(2, "0");
    const seconds = date.getSeconds().toString().padStart(2, "0");
    const ampm = hours >= 12 ? "PM" : "AM";
    hours = hours % 12;
    hours = hours ? hours : 12; // Convert 0 to 12

    // Calculate relative time
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    let relativeTime = "";
    if (diffMs < 0) {
      // Future date
      relativeTime = "In the future";
    } else if (diffDay > 365) {
      relativeTime = `${Math.floor(diffDay / 365)} year${
        Math.floor(diffDay / 365) > 1 ? "s" : ""
      } ago`;
    } else if (diffDay > 30) {
      relativeTime = `${Math.floor(diffDay / 30)} month${
        Math.floor(diffDay / 30) > 1 ? "s" : ""
      } ago`;
    } else if (diffDay > 0) {
      relativeTime = `${diffDay} day${diffDay > 1 ? "s" : ""} ago`;
    } else if (diffHour > 0) {
      relativeTime = `${diffHour} hour${diffHour > 1 ? "s" : ""} ago`;
    } else if (diffMin > 0) {
      relativeTime = `${diffMin} minute${diffMin > 1 ? "s" : ""} ago`;
    } else {
      relativeTime = "Just now";
    }

    // Full date for tooltip (in local timezone)
    const fullDate = `${year}-${(date.getMonth() + 1)
      .toString()
      .padStart(2, "0")}-${day.toString().padStart(2, "0")} ${hours
      .toString()
      .padStart(2, "0")}:${minutes}:${seconds} ${ampm}`;

    return {
      date: `${month} ${day}, ${year}`,
      time: `${hours}:${minutes} ${ampm}`,
      fullDate,
      relativeTime,
    };
  } catch (error) {
    return { date: "Invalid date", time: "", fullDate: "", relativeTime: "" };
  }
}
