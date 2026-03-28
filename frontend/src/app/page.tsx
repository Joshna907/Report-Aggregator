"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { Lock, User, Shield } from "lucide-react";

export default function Page() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    // Simulate auth delay
    setTimeout(() => {
      setLoading(false);
      router.push("/dashboard");
    }, 800);
  };

  return (
    <div className="fixed inset-0 z-50 bg-[#f8fafc] flex flex-col items-center justify-center font-sans">
      
      {/* Login Card */}
      <div className="w-full max-w-md bg-white border border-gray-200 shadow-lg rounded-lg overflow-hidden flex flex-col">
        
        {/* Header Section */}
        <div className="bg-[#004481] p-6 text-center text-white border-b-4 border-foss-red">
          <div className="flex justify-center mb-3">
             <Shield size={40} className="text-white opacity-90" />
          </div>
          <h1 className="text-2xl font-bold tracking-wider">
            Report<span className="text-[#99ccff]">Aggregator</span>
          </h1>
          <p className="text-xs text-gray-200 mt-2 font-medium uppercase tracking-widest">Compliance Harmonization Engine</p>
        </div>

        {/* Form Section */}
        <div className="p-8">
          <div className="flex bg-blue-50 text-[#004481] p-3 text-xs mb-6 rounded border border-blue-100 items-start gap-2">
             <Lock size={16} className="mt-0.5 flex-shrink-0" />
             <p>Authorization Required. Use your FOSSology credentials to access the aggregation results.</p>
          </div>

          <form onSubmit={handleLogin} className="space-y-5">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-1">Username</label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400">
                  <User size={16} />
                </div>
                <input 
                  type="text" 
                  defaultValue="fossy"
                  required
                  className="w-full border border-gray-300 rounded pl-10 pr-3 py-2 text-sm focus:border-[#004481] focus:ring-1 focus:ring-[#004481] outline-none transition-all" 
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-1">Password</label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400">
                  <Lock size={16} />
                </div>
                <input 
                  type="password" 
                  defaultValue="admin"
                  required
                  className="w-full border border-gray-300 rounded pl-10 pr-3 py-2 text-sm focus:border-[#004481] focus:ring-1 focus:ring-[#004481] outline-none transition-all" 
                />
              </div>
            </div>

            <div className="pt-2">
              <button 
                type="submit" 
                disabled={loading}
                className="w-full bg-[#004481] hover:bg-[#003366] text-white py-2.5 rounded font-bold shadow-sm transition-colors disabled:opacity-70 flex justify-center items-center h-[44px]"
              >
                {loading ? (
                  <div className="h-5 w-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : (
                  "Log In"
                )}
              </button>
            </div>
          </form>
        </div>

        {/* Footer */}
        <div className="bg-gray-50 border-t border-gray-200 px-8 py-4 text-center">
          <p className="text-xs text-gray-500">
            Forgot your password? Contact your FOSSology administrator.
          </p>
        </div>

      </div>

      <div className="mt-8 text-xs text-gray-400">
        &copy; 2026 Internal FOSSology &amp; SW360 Aggregation Hub
      </div>

    </div>
  );
}
