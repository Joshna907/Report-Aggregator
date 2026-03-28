import Link from "next/link";
import React from "react";

export function TopNav() {
  return (
    <header className="bg-[#004481] text-white shadow-md w-full">
      <div className="max-w-[1400px] mx-auto flex items-center justify-between px-6 py-3">
        
        {/* Logo left */}
        <div className="flex items-center space-x-8">
          <div className="text-xl font-bold tracking-wider">
            Report<span className="text-[#99ccff]">Aggregator</span>
          </div>

          {/* Nav Links */}
          <nav className="hidden xl:flex space-x-6 text-sm font-medium">
            <Link href="/dashboard" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Dashboard</Link>
            <Link href="/upload" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Upload & Merge</Link>
            <Link href="/browse" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Repository Browser</Link>
            <Link href="/conflicts" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Conflict Manager</Link>
            <Link href="/export" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Export & Sync</Link>
            <Link href="/jobs" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">FOSSology Bridge</Link>
            <Link href="/changelog" className="hover:text-[#99ccff] transition-colors border-b-2 border-transparent hover:border-[#99ccff] pb-1">Audit Trail</Link>
          </nav>
        </div>

        {/* User context right */}
        <div className="flex items-center space-x-6">
          <nav className="hidden lg:flex space-x-4 text-xs font-semibold text-gray-200">
            <Link href="/admin" className="hover:text-white uppercase tracking-tighter">Admin</Link>
            <Link href="/help" className="hover:text-white uppercase tracking-tighter">Help</Link>
          </nav>

          <div className="h-6 w-px bg-[#336699]"></div>

          <div className="flex items-center space-x-4 text-xs">
            <div className="text-right leading-tight hidden lg:block text-gray-200">
              <div>User: <span className="font-semibold text-white">fossy</span></div>
              <div>Group: <span className="font-semibold text-white">admin</span></div>
            </div>
            <Link href="/">
              <button className="bg-[#003366] hover:bg-[#b91c1c] border border-[#336699] px-3 py-1.5 rounded transition-all text-white font-bold shadow-sm uppercase tracking-tighter text-[10px]">
                Logout
              </button>
            </Link>
          </div>
        </div>

      </div>
    </header>
  );
}
