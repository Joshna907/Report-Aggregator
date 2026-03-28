"use client";

import React from "react";
import { HelpCircle, FileText, Github, Globe, MessageSquare, BookOpen } from "lucide-react";

export default function HelpPage() {
  return (
    <div className="space-y-6 animate-in fade-in duration-500 max-w-4xl mx-auto">
      <div className="text-center py-8">
        <div className="inline-flex p-3 bg-blue-50 rounded-full text-foss-blue mb-4">
            <HelpCircle size={48} />
        </div>
        <h1 className="text-3xl font-bold text-gray-800">Support & Documentation</h1>
        <p className="text-gray-500 mt-2">Everything you need to master the FOSSology Report Aggregator.</p>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-md transition-shadow cursor-pointer group">
            <div className="flex items-start gap-4">
                <div className="p-2 bg-gray-50 rounded text-gray-600 group-hover:bg-foss-blue group-hover:text-white transition-colors">
                    <BookOpen size={24} />
                </div>
                <div className="space-y-1">
                    <h2 className="font-bold text-gray-800">User Handbook</h2>
                    <p className="text-sm text-gray-500">Learn how to merge SPDX 2.3 and CycloneDX formats effectively.</p>
                </div>
            </div>
        </div>
        <div className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-md transition-shadow cursor-pointer group">
            <div className="flex items-start gap-4">
                <div className="p-2 bg-gray-50 rounded text-gray-600 group-hover:bg-foss-blue group-hover:text-white transition-colors">
                    <FileText size={24} />
                </div>
                <div className="space-y-1">
                    <h2 className="font-bold text-gray-800">API Documentation</h2>
                    <p className="text-sm text-gray-500">Integrate the aggregator into your CI/CD pipeline via REST.</p>
                </div>
            </div>
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-100 rounded-lg p-8 space-y-4">
        <h2 className="text-lg font-bold text-blue-900 flex items-center gap-2">
            <MessageSquare size={20} />
            Community & GSoC 2024
        </h2>
        <p className="text-sm text-blue-800 leading-relaxed">
            This project is part of the FOSSology ecosystem. If you are a GSoC contributor or a security professional using this for compliance, we value your feedback! Join our community channels to discuss <strong>SPDX 3.0</strong> support and future <strong>VEX</strong> integration.
        </p>
        <div className="flex gap-4 pt-2">
            <a href="https://github.com/fossology" target="_blank" rel="noreferrer" className="flex items-center gap-2 text-xs font-bold text-foss-blue bg-white px-4 py-2 rounded border border-blue-200 hover:bg-blue-600 hover:text-white transition-all">
                <Github size={14} />
                GitHub Repository
            </a>
            <a href="https://www.fossology.org" target="_blank" rel="noreferrer" className="flex items-center gap-2 text-xs font-bold text-foss-blue bg-white px-4 py-2 rounded border border-blue-200 hover:bg-blue-600 hover:text-white transition-all">
                <Globe size={14} />
                Official Foundation Site
            </a>
        </div>
      </div>

      <div className="py-8 border-t border-gray-100 text-center space-y-2">
        <p className="text-xs text-gray-400">© 2024 FOSSology Foundation • GSoC Refinement Package</p>
        <p className="text-[10px] text-gray-300">Version 1.0.4-rc (Aggregator Engine v0.9.1)</p>
      </div>
    </div>
  );
}
