"use client";

import { useEffect, Suspense } from "react";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";

function LoginRedirect() {
  const searchParams = useSearchParams();
  const callbackUrl = searchParams?.get("next") || "/customer";

  useEffect(() => {
    // Automatically redirect to Keycloak and skip the intermediate button page
    signIn("keycloak", { callbackUrl });
  }, [callbackUrl]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 p-6 dark:bg-slate-950">
      <div className="text-center">
        <p className="text-sm text-slate-500 dark:text-slate-400">
          Redirecting to Ethiopian Insurance Corporation secure login...
        </p>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <LoginRedirect />
    </Suspense>
  );
}
