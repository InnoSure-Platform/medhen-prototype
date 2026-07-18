"use client";

import { useEffect, Suspense } from "react";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";

function LoginRedirect() {
  const searchParams = useSearchParams();
  const callbackUrl = searchParams?.get("next") || "/customer";
  const error = searchParams?.get("error");

  useEffect(() => {
    // Only automatically redirect if there is no error.
    // If there is an error, we stop the loop and display it.
    if (!error) {
      signIn("keycloak", { callbackUrl });
    }
  }, [callbackUrl, error]);

  if (error) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-slate-50 p-6 dark:bg-slate-950">
        <div className="text-center max-w-md rounded-xl bg-white p-8 shadow-sm dark:bg-slate-900 border border-red-200 dark:border-red-900">
          <h2 className="text-xl font-bold text-red-600 mb-2">Authentication Error</h2>
          <p className="text-slate-600 dark:text-slate-300 mb-6">
            There was a problem authenticating with the identity provider. 
            (Error code: {error})
          </p>
          <a href="/" className="px-4 py-2 bg-brand-blue-600 text-white rounded-md hover:bg-brand-blue-700">
            Return to Home
          </a>
        </div>
      </div>
    );
  }

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
