"use client";

// Catches errors in the root layout itself; must render its own <html>/<body>.
export default function GlobalError({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <html lang="en">
      <body style={{ display: "grid", placeItems: "center", minHeight: "100dvh", fontFamily: "system-ui, sans-serif" }}>
        <div style={{ textAlign: "center", maxWidth: "28rem", padding: "2rem" }}>
          <h1 style={{ fontSize: "1.5rem", fontWeight: 700 }}>Something went wrong</h1>
          <p style={{ color: "#64748b", marginTop: "0.5rem" }}>An unexpected error occurred.</p>
          <button
            onClick={reset}
            style={{ marginTop: "1rem", padding: "0.6rem 1.25rem", borderRadius: "0.75rem", background: "#2563eb", color: "#fff", border: 0, fontWeight: 600 }}
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
