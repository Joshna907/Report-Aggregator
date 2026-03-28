"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { FolderUp, GitMerge, AlertCircle, FileText } from "lucide-react";
import axios from "axios";
import API_BASE from "@/lib/api";

export default function HomePage() {
  const [stats, setStats] = useState({ components: 0, conflicts: 0, reports: 0 });

  useEffect(() => {
    // Attempt to fetch real stats if backend is connected
    axios.get(`${API_BASE}/summary`)
      .then(res => {
        setStats({
          components: res.data.totalComponents || 0,
          conflicts: res.data.totalConflicts || 0,
          reports: res.data.sourceReportsCount || 2 // Fallback if missing
        });
      })
      .catch(() => console.log("Backend not reachable for initial stats yet."));
  }, []);

  return (
    <div className="font-sans text-gray-800">
      
      {/* Header Section */}
      <div className="flex justify-between items-end mb-8 border-b border-gray-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold text-[#004481] m-0">Project Dashboard</h1>
          <p className="text-sm text-gray-500 mt-1">
            Report Aggregator • v1.0.0-rc1 • Standalone Prototype
          </p>
        </div>
        <Link href="/upload">
          <button className="bg-[#004481] hover:bg-[#003366] text-white px-5 py-2 rounded-md font-semibold shadow transition-all flex items-center space-x-2">
            <FolderUp size={18} />
            <span>New Merge</span>
          </button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        
        {/* Stat Card 1 */}
        <div className="bg-white border text-center border-gray-200 rounded-lg shadow-sm p-6 flex flex-col items-center">
          <div className="rounded-full bg-blue-50 p-3 mb-3 text-[#004481]">
            <FileText size={28} />
          </div>
          <h3 className="text-3xl font-bold text-gray-800">{stats.components === 0 ? "..." : stats.components}</h3>
          <p className="text-sm text-gray-500 font-medium uppercase tracking-wide">Aggregated Components</p>
        </div>

        {/* Stat Card 2 */}
        <div className="bg-white border text-center border-gray-200 rounded-lg shadow-sm p-6 flex flex-col items-center">
          <div className="rounded-full bg-red-50 p-3 mb-3 text-[#c12127]">
            <AlertCircle size={28} />
          </div>
          <h3 className="text-3xl font-bold text-gray-800">{stats.conflicts}</h3>
          <p className="text-sm text-gray-500 font-medium uppercase tracking-wide">Open Conflicts</p>
        </div>

        {/* Stat Card 3 */}
        <div className="bg-white border text-center border-gray-200 rounded-lg shadow-sm p-6 flex flex-col items-center">
          <div className="rounded-full bg-green-50 p-3 mb-3 text-green-700">
            <GitMerge size={28} />
          </div>
          <h3 className="text-3xl font-bold text-gray-800">{stats.components === 0 ? "..." : stats.reports}</h3>
          <p className="text-sm text-gray-500 font-medium uppercase tracking-wide">Source Reports Processed</p>
        </div>

      </div>

      {/* Main Content Info */}
      <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden flex flex-col md:flex-row">
        <div className="p-8 md:w-2/3 border-r border-gray-200">
          <h2 className="text-xl font-bold text-gray-800 mb-4 flex items-center">
            Getting Started
          </h2>
          <p className="mb-4 text-gray-600 leading-relaxed text-sm">
            This tool unifies multiple scan outputs (SPDX, CycloneDX) into a single aggregated report. 
            It identifies duplicate components using Package URLs and manages metadata conflicts between different scanners.
          </p>
          <ul className="list-inside list-disc text-sm text-gray-600 space-y-2 mb-6">
            <li>Upload multiple SBOM files into the secure repository.</li>
            <li>Detect identical components across reports using Package URLs.</li>
            <li>Identify and flag conflicting metadata (Licenses, Copyrights).</li>
            <li>Export the finalized report to SW360.</li>
          </ul>
        </div>
        
        <div className="bg-[#f8fafc] p-8 md:w-1/3 flex flex-col justify-center">
           <h3 className="text-sm font-bold uppercase tracking-wider text-gray-500 mb-4">Quick Actions</h3>
           <div className="space-y-3">
             <Link href="/browse" className="block w-full bg-white border border-gray-200 hover:border-[#004481] hover:text-[#004481] transition-colors p-3 rounded text-sm font-semibold text-center text-gray-700 shadow-sm">
               Browse Aggregated Data
             </Link>
             <Link href="/conflicts" className="block w-full bg-white border border-gray-200 hover:border-[#c12127] hover:text-[#c12127] transition-colors p-3 rounded text-sm font-semibold text-center text-gray-700 shadow-sm">
               Resolve Conflicts
             </Link>
           </div>
        </div>
      </div>

    </div>
  );
}
