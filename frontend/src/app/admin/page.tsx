"use client";

import React from "react";
import { Settings, User, Database, Globe, Save } from "lucide-react";

export default function AdminPage() {
  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">System Administration</h1>
        <p className="text-sm text-gray-500 mt-1">Configure global aggregator settings and external API credentials.</p>
      </div>

      <div className="grid grid-cols-4 gap-6">
        {/* Sidebar settings */}
        <div className="col-span-1 space-y-1">
          {[
            { name: "General Settings", icon: Settings, active: true },
            { name: "User Management", icon: User, active: false },
            { name: "Database & Backup", icon: Database, active: false },
            { name: "Network & APIs", icon: Globe, active: false },
          ].map((item) => (
            <button key={item.name} className={`w-full flex items-center gap-3 px-4 py-2 text-sm rounded transition-colors ${item.active ? 'bg-foss-blue text-white' : 'text-gray-600 hover:bg-gray-100'}`}>
              <item.icon size={16} />
              {item.name}
            </button>
          ))}
        </div>

        {/* Main settings area */}
        <div className="col-span-3 bg-white border border-gray-200 rounded p-8 space-y-8">
          <section className="space-y-4">
              <h2 className="text-lg font-bold text-gray-800 border-b border-gray-100 pb-2">Global Merge Configuration</h2>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-gray-600 uppercase">Conflict Sensitivity</label>
                  <select className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:ring-1 focus:ring-foss-blue outline-none transition-all">
                    <option>High (Flag everything)</option>
                    <option defaultValue="normal">Standard (Merge identicals)</option>
                    <option>Low (Auto-resolve common)</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-bold text-gray-600 uppercase">Default PURL Format</label>
                  <select className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:ring-1 focus:ring-foss-blue outline-none transition-all">
                    <option>Canonical PURL</option>
                    <option>Versioned PURL</option>
                    <option>Hash-Only</option>
                  </select>
                </div>
              </div>
          </section>

          <section className="space-y-4">
              <h2 className="text-lg font-bold text-gray-800 border-b border-gray-100 pb-2">External API Integration</h2>
              <div className="space-y-4">
                <div className="space-y-1.5">
                    <label className="text-xs font-bold text-gray-600 uppercase">FOSSology REST Endpoint</label>
                    <input type="text" defaultValue="http://localhost:8081/api/v2" className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:ring-1 focus:ring-foss-blue outline-none transition-all" />
                </div>
                <div className="space-y-1.5">
                    <label className="text-xs font-bold text-gray-600 uppercase">SW360 API Token (Secret)</label>
                    <input type="password" value="************************" readOnly className="w-full border border-gray-300 rounded px-3 py-2 text-sm bg-gray-50 focus:ring-1 focus:ring-foss-blue outline-none transition-all" />
                </div>
              </div>
          </section>

          <div className="flex justify-end border-t border-gray-100 pt-6">
            <button className="bg-[#004481] hover:bg-[#003366] text-white font-bold py-2 px-6 rounded transition-all shadow-md active:scale-[0.98] flex items-center gap-2">
                <Save size={16} />
                Save All Changes
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
