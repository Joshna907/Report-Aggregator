"use client";

import React, { useEffect, useState } from "react";
import axios from "axios";
import API_BASE from "@/lib/api";
import { Server, DownloadCloud, FileCode, CheckCircle2, RotateCw, AlertTriangle } from "lucide-react";
import { useRouter } from "next/navigation";

interface FossologyUpload {
  id: number;
  uploadName: string;
  uploadDate: string;
  folderName: string;
}

export default function JobsPage() {
  const [uploads, setUploads] = useState<FossologyUpload[]>([]);
  const [loading, setLoading] = useState(true);
  const [fetching, setFetching] = useState<number | null>(null);
  const router = useRouter();

  useEffect(() => {
    fetchUploads();
  }, []);

  const fetchUploads = () => {
    setLoading(true);
    axios.get(`${API_BASE}/fossology/uploads`)
      .then(res => {
        setUploads(res.data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  };

  const handleFetchReport = async (uploadId: number, format: string) => {
    setFetching(uploadId);
    try {
      const res = await axios.post(`${API_BASE}/fossology/fetch`, {
        uploadId,
        format
      });
      alert(`Success: ${res.data.message}\nFound ${res.data.components} components.`);
      router.push("/upload"); // Go to upload page to see it in the list or merge
    } catch (err: any) {
      alert("Fetch failed: " + (err.response?.data?.error || err.message));
    } finally {
      setFetching(null);
    }
  };

  return (
    <div className="flex flex-col font-sans text-gray-800 max-w-6xl mx-auto py-8">
      <div className="mb-8 border-b border-gray-200 pb-6 flex justify-between items-end">
        <div>
          <h1 className="text-3xl font-bold text-[#004481] flex items-center gap-3">
            <Server size={32} />
            FOSSology Report Bridge
          </h1>
          <p className="text-sm text-gray-500 mt-2">
            Pull scan reports directly from your FOSSology instance into the aggregator engine.
          </p>
        </div>
        <button 
          onClick={fetchUploads} 
          className="p-2 text-gray-400 hover:text-[#004481] transition-colors"
        >
          <RotateCw size={20} className={loading ? "animate-spin" : ""} />
        </button>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-8 flex items-start gap-4 shadow-sm">
        <DownloadCloud className="text-[#004481] shrink-0" size={24} />
        <div>
          <h3 className="font-bold text-[#004481]">Direct Integration Ready</h3>
          <p className="text-sm text-blue-800 leading-relaxed mt-1">
            The bridge is connected to <strong>localhost:8081</strong>. You can trigger report generations and ingest the outputs as SPDX or CycloneDX without manual export/upload cycles.
          </p>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-20 text-gray-500">
          <RotateCw className="animate-spin mx-auto mb-4" size={32} />
          Fetching upload list from FOSSology...
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {uploads.map((u) => (
            <div key={u.id} className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden flex flex-col group hover:border-[#004481] transition-all">
              <div className="p-5 flex-1">
                <div className="flex justify-between items-start mb-3">
                  <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest bg-gray-50 px-2 py-1 rounded">
                    Upload ID: #{u.id}
                  </span>
                  <span className="text-[10px] font-bold text-blue-600 bg-blue-50 px-2 py-1 rounded border border-blue-100 uppercase">
                    {u.folderName}
                  </span>
                </div>
                <h3 className="font-bold text-gray-800 mb-2 group-hover:text-[#004481] transition-colors line-clamp-2">
                  {u.uploadName}
                </h3>
                <p className="text-xs text-gray-500 flex items-center gap-1.5 mt-auto">
                  Scan completed: {u.uploadDate}
                </p>
              </div>
              
              <div className="bg-gray-50 p-4 border-t border-gray-100 grid grid-cols-2 gap-3">
                <button 
                  onClick={() => handleFetchReport(u.id, "spdx")}
                  disabled={!!fetching}
                  className="bg-white border border-gray-300 hover:border-[#004481] hover:text-[#004481] text-gray-700 py-2 rounded text-[11px] font-bold flex items-center justify-center gap-1.5 transition-all shadow-sm active:scale-95 disabled:opacity-50"
                >
                  <FileCode size={14} />
                  {fetching === u.id ? "Syncing..." : "Fetch SPDX"}
                </button>
                <button 
                  onClick={() => handleFetchReport(u.id, "cyclonedx")}
                  disabled={!!fetching}
                  className="bg-white border border-gray-300 hover:border-[#004481] hover:text-[#004481] text-gray-700 py-2 rounded text-[11px] font-bold flex items-center justify-center gap-1.5 transition-all shadow-sm active:scale-95 disabled:opacity-50"
                >
                  <FileCode size={14} />
                  {fetching === u.id ? "Syncing..." : "Fetch CDX"}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="mt-12 p-6 bg-gray-50 border border-gray-200 rounded-xl flex items-center gap-4">
        <AlertTriangle className="text-amber-500" size={24} />
        <p className="text-xs text-gray-500 italic">
          <strong>Note:</strong> Report generation in FOSSology can take time depending on the project size. If "Fetch" fails, check the FOSSology UI for job status.
        </p>
      </div>
    </div>
  );
}
