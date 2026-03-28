"use client";

import React, { useState } from "react";
import axios from "axios";
import API_BASE from "@/lib/api";
import { Download, ExternalLink, ShieldCheck, History, Clock, CheckCircle2, FileJson, FileText } from "lucide-react";

export default function ExportPage() {
  const [loading, setLoading] = useState(false);
  const [downloadingFormat, setDownloadingFormat] = useState<string | null>(null);
  const [lastExport, setLastExport] = useState<any>(null);

  const handleExportSw360 = async () => {
    setLoading(true);
    try {
      const res = await axios.post(`${API_BASE}/result/export/sw360`);
      setLastExport({
        timestamp: new Date().toLocaleString(),
        project: "FOSS-AGGREGATE-PROJ",
        status: "Success",
        details: res.data.message,
        type: "SW360 Push"
      });
    } catch (err: any) {
      alert("Export failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(true);
      setTimeout(() => setLoading(false), 500);
    }
  };

  const handleDownloadReport = async (format: string) => {
    setDownloadingFormat(format);
    try {
      const res = await axios.get(`${API_BASE}/export?format=${format}`, {
        responseType: "blob",
      });
      const blob = new Blob([res.data], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      const ext = format === "cyclonedx" ? "cdx.json" : "spdx.json";
      link.download = `aggregated-report-${new Date().toISOString().split("T")[0]}.${ext}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      setLastExport({
        timestamp: new Date().toLocaleString(),
        project: format === "spdx" ? "SPDX 2.3 Report" : "CycloneDX Report",
        status: "Downloaded",
        details: `Exported merged result as ${format.toUpperCase()}`,
        type: `${format.toUpperCase()} Download`
      });
    } catch (err: any) {
      alert("Download failed: " + (err.response?.data?.error || err.message));
    } finally {
      setDownloadingFormat(null);
    }
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Export Management</h1>
          <p className="text-sm text-gray-500 mt-1">Download aggregated reports or synchronize with external ecosystems.</p>
        </div>
      </div>

      {/* Download Reports Section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        
        {/* SPDX 2.3 Download Card */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="bg-[#1a6b3c] p-4 text-white flex items-center gap-2">
            <FileJson size={18} />
            <h2 className="font-semibold text-lg">SPDX 2.3 Export</h2>
          </div>
          <div className="p-6 space-y-4">
            <p className="text-sm text-gray-600">
              Download the aggregated report as a valid SPDX 2.3 JSON document. Includes all merged components with licenses, hashes, suppliers, and provenance data.
            </p>
            <div className="bg-green-50 p-3 rounded border border-green-100 flex items-start gap-3">
              <ShieldCheck className="text-green-700 mt-0.5" size={18} />
              <div className="text-xs text-green-800">
                <p className="font-semibold">Standards Compliant</p>
                <p>Output conforms to SPDX 2.3 specification using spdx/tools-golang.</p>
              </div>
            </div>
            <button 
              onClick={() => handleDownloadReport("spdx")}
              disabled={downloadingFormat === "spdx"}
              className="w-full bg-[#1a6b3c] hover:bg-[#155a31] text-white font-bold py-3 px-4 rounded transition-all shadow-md active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <Download size={16} />
              {downloadingFormat === "spdx" ? "Generating..." : "Download SPDX 2.3 JSON"}
            </button>
          </div>
        </div>

        {/* CycloneDX Download Card */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="bg-[#6f42c1] p-4 text-white flex items-center gap-2">
            <FileText size={18} />
            <h2 className="font-semibold text-lg">CycloneDX Export</h2>
          </div>
          <div className="p-6 space-y-4">
            <p className="text-sm text-gray-600">
              Download the aggregated report as a CycloneDX BOM JSON document. Includes all merged components with PURL identifiers, licenses, and hash verification data.
            </p>
            <div className="bg-purple-50 p-3 rounded border border-purple-100 flex items-start gap-3">
              <ShieldCheck className="text-purple-700 mt-0.5" size={18} />
              <div className="text-xs text-purple-800">
                <p className="font-semibold">BOM Standard</p>
                <p>Output uses CycloneDX Go library for spec-compliant encoding.</p>
              </div>
            </div>
            <button 
              onClick={() => handleDownloadReport("cyclonedx")}
              disabled={downloadingFormat === "cyclonedx"}
              className="w-full bg-[#6f42c1] hover:bg-[#5a32a3] text-white font-bold py-3 px-4 rounded transition-all shadow-md active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <Download size={16} />
              {downloadingFormat === "cyclonedx" ? "Generating..." : "Download CycloneDX JSON"}
            </button>
          </div>
        </div>
      </div>

      {/* SW360 Integration */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="bg-foss-blue p-4 text-white flex items-center gap-2">
            <ExternalLink size={18} />
            <h2 className="font-semibold text-lg">SW360 Integration</h2>
          </div>
          <div className="p-6 space-y-4">
            <p className="text-sm text-gray-600">
              Push your authoritative SBOM result directly to the SW360 Component Catalog for full lifecycle management.
            </p>
            <button 
              onClick={handleExportSw360}
              disabled={loading}
              className="w-full bg-[#004481] hover:bg-[#003366] text-white font-bold py-3 px-4 rounded transition-all shadow-md active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {loading ? "Synchronizing..." : "Execute SW360 Push"}
            </button>
          </div>
        </div>

        {/* Recent Exports */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 flex flex-col">
          <div className="p-4 border-b border-gray-100 flex items-center gap-2 text-gray-700 font-semibold">
            <History size={18} />
            <h2>Recent Exports</h2>
          </div>
          <div className="flex-1 p-6 flex flex-col items-center justify-center text-center space-y-3">
            {lastExport ? (
              <div className="w-full space-y-4">
                <div className="flex items-center justify-between p-3 bg-green-50 rounded border border-green-100">
                  <div className="flex items-center gap-3">
                    <CheckCircle2 className="text-green-600" size={20} />
                    <div className="text-left">
                      <p className="text-sm font-bold text-green-800">{lastExport.type}</p>
                      <p className="text-xs text-green-700">{lastExport.timestamp}</p>
                    </div>
                  </div>
                  <span className="text-[10px] bg-green-200 text-green-800 px-2 py-0.5 rounded uppercase font-bold tracking-tighter">Done</span>
                </div>
                <div className="text-left text-xs text-gray-600 space-y-1">
                  <p><strong>Target:</strong> {lastExport.project}</p>
                  <p><strong>Response:</strong> {lastExport.details}</p>
                </div>
              </div>
            ) : (
              <>
                <div className="p-4 bg-gray-50 rounded-full text-gray-300">
                  <Clock size={48} />
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-600">No export history found</p>
                  <p className="text-xs text-gray-400">Download a report or push to SW360 to see logs here.</p>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
