"use client";

import { Suspense } from "react";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";
import Image from "next/image";

function LoginContent() {
  const searchParams = useSearchParams();
  const callbackUrl = searchParams?.get("next") || "/customer";
  const error = searchParams?.get("error");

  const handleLogin = () => {
    signIn("keycloak", { callbackUrl });
  };

  return (
    <div className="min-h-screen w-full flex flex-col md:flex-row bg-white">
      {/* Left side - Cinematic Brand Visual */}
      <div className="relative hidden md:flex flex-col flex-1 items-center justify-center bg-brand-blue-900 overflow-hidden">
        {/* Dynamic Background Gradients */}
        <div className="absolute inset-0 bg-gradient-to-br from-brand-blue-900 via-brand-blue-600 to-brand-blue-900 opacity-90" />
        <div className="absolute -top-[20%] -left-[10%] w-[70%] h-[70%] rounded-full bg-brand-blue-500/20 blur-3xl mix-blend-screen" />
        <div className="absolute bottom-[10%] right-[10%] w-[50%] h-[50%] rounded-full bg-brand-gold/10 blur-3xl mix-blend-screen" />

        <div className="relative z-10 p-16 max-w-2xl text-center">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-white/10 backdrop-blur-md border border-white/20 mb-8 text-white/80 text-sm font-medium tracking-wide">
            <span className="w-2 h-2 rounded-full bg-brand-gold animate-pulse" />
            Enterprise Tier-0 Portal
          </div>
          <h1 className="text-4xl lg:text-6xl font-bold text-white mb-6 leading-tight">
            Secure Infrastructure for a Modern Era
          </h1>
          <p className="text-lg text-white/70 leading-relaxed max-w-lg mx-auto">
            Ethiopian Insurance Corporation's next-generation platform for brokers, underwriters, and customers.
          </p>
        </div>
      </div>

      {/* Right side - Authentication Card */}
      <div className="flex-1 flex items-center justify-center p-8 bg-slate-50">
        <div className="w-full max-w-md bg-white p-10 rounded-2xl shadow-xl border border-slate-200/60 relative">
          
          <div className="text-center mb-10">
            <h2 className="text-3xl font-bold text-slate-900 mb-2 font-display">Welcome Back</h2>
            <p className="text-slate-500">Sign in to access your dashboard</p>
          </div>

          {error && (
            <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 text-sm text-red-600">
              <span className="font-semibold block mb-1">Authentication Error</span>
              There was a problem authenticating with the identity provider. (Error code: {error})
            </div>
          )}

          <div className="space-y-6">
            <button
              onClick={handleLogin}
              className="w-full flex items-center justify-center gap-3 bg-brand-blue-600 hover:bg-brand-blue-500 text-white font-semibold py-3.5 px-4 rounded-xl transition-all duration-300 shadow-md shadow-brand-blue-600/20 hover:shadow-lg hover:shadow-brand-blue-600/30 active:scale-[0.98]"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              Sign In with EIC SSO
            </button>
            
            <p className="text-xs text-center text-slate-400 mt-6 leading-relaxed">
              By signing in, you agree to the <a href="#" className="text-brand-blue-600 hover:underline">Terms of Service</a> and <a href="#" className="text-brand-blue-600 hover:underline">Privacy Policy</a>.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-blue-600"></div>
      </div>
    }>
      <LoginContent />
    </Suspense>
  );
}
