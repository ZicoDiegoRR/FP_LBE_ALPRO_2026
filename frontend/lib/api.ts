// Points at the Go/Gin backend. Override with NEXT_PUBLIC_API_BASE_URL in
// .env.local if your backend runs on a different host/port.
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
