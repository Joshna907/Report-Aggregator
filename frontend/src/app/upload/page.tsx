"use client";

import React, { useState } from "react";
import axios from "axios";
import { useRouter } from "next/navigation";
import API_BASE from "@/lib/api";
import { UploadCloud, CheckCircle2 } from "lucide-react";

export default function UploadPage() {
  const [files, setFiles] = useState<File[]>([]);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      setFiles(prev => {
        // Filter out duplicates by name+size to avoid confusion
        const combined = [...prev, ...newFiles];
        const unique = combined.filter((file, index, self) =>
          index === self.findIndex((f) => f.name === file.name && f.size === file.size)
        );
        return unique;
      });
      // Reset input value so the same file can be picked again if removed
      e.target.value = "";
    }
  };

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index));
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (files.length === 0) return;

    setLoading(true);
    const formData = new FormData();
    files.forEach((f) => formData.append("files", f));

    try {
      await axios.post(`${API_BASE}/merge`, formData, {
        headers: { "Content-Type": "multipart/form-data" }
      });
      router.push("/browse");
    } catch (err) {
      console.error(err);
      alert("Upload failed. Is the backend running at :8080?");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col font-sans text-gray-800 max-w-5xl mx-auto">
      <div className="mb-6 border-b border-gray-200 pb-4">
        <h1 className="text-2xl font-bold text-[#004481] flex items-center gap-2">
          <UploadCloud className="text-[#004481]" />
          Upload Sub-BOMs for Aggregation
        </h1>
        <p className="text-sm text-gray-500 mt-2 max-w-3xl">
          Securely upload multiple SBOM files (SPDX, CycloneDX) to the aggregator. The system will parse formats, deduplicate identical components, preserve hierarchical relationships, and flag metadata conflicts.
        </p>
      </div>

      <form onSubmit={handleUpload} className="bg-white border text-sm max-w-4xl border-gray-200 shadow-sm rounded-lg p-8 space-y-8">
        
        {/* Step 1 */}
        <div className="flex flex-col space-y-2">
          <h3 className="font-bold text-gray-800 border-b border-gray-100 pb-2">1. Destination Context</h3>
          <p className="text-xs text-gray-500 mb-2">Select the repository folder where the aggregated result will be stored.</p>
          <select className="border border-gray-300 rounded p-2 form-select w-72 bg-gray-50 text-gray-700">
            <option>Software Repository / Main Project</option>
            <option>Vendor Audits</option>
          </select>
        </div>

        {/* Step 2 */}
        <div className="flex flex-col space-y-2">
          <h3 className="font-bold text-gray-800 border-b border-gray-100 pb-2">2. Select Files</h3>
          <p className="text-xs text-gray-500 mb-2">You can select multiple JSON or XML files at once.</p>
          <div className="flex flex-col space-y-3">
            <input 
              type="file" 
              multiple 
              onChange={handleFileChange}
              className="border border-dashed py-6 text-center px-4 border-gray-400 bg-gray-50 rounded cursor-pointer text-gray-700 hover:bg-gray-100 transition-colors"
            />
            {files.length > 0 && (
              <div className="bg-blue-50 p-4 border border-blue-100 rounded text-sm">
                <strong className="text-[#004481] flex items-center gap-2 mb-2">
                  <CheckCircle2 size={16} /> Selected files ready for merge:
                </strong>
                <ul className="space-y-2 mt-4">
                  {files.map((f, idx) => (
                    <li key={`${f.name}-${idx}`} className="flex items-center justify-between bg-white/60 p-2 rounded-md border border-blue-100 shadow-sm">
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-medium text-gray-700 truncate">{f.name}</span>
                        <span className="text-[10px] text-gray-500 bg-white px-2 py-0.5 rounded border border-gray-100">
                          {Math.round(f.size / 1024)} KB
                        </span>
                      </div>
                      <button 
                        type="button"
                        onClick={() => removeFile(idx)}
                        className="text-gray-400 hover:text-red-500 transition-colors p-1"
                        title="Remove file"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>

        {/* Step 3 */}
        <div className="flex flex-col space-y-2">
          <h3 className="font-bold text-gray-800 border-b border-gray-100 pb-2">3. Merge Description (Optional)</h3>
          <input type="text" className="border border-gray-300 rounded p-2 form-input w-full bg-gray-50 text-gray-700" placeholder="e.g. Initial merge of vendor SBOMs for Q3 Release" />
        </div>

        {/* Step 4 */}
        <div className="flex flex-col space-y-2">
          <h3 className="font-bold text-gray-800 border-b border-gray-100 pb-2">4. Pipeline Analysis Options</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
            <label className="flex items-start space-x-3 p-3 border border-gray-200 rounded bg-gray-50">
              <input type="checkbox" disabled checked className="mt-1" /> 
              <span className="text-gray-700 leading-tight"><strong>Format Normalization</strong><br/><span className="text-xs text-gray-500">Converts SPDX/CDX to unified models.</span></span>
            </label>
            <label className="flex items-start space-x-3 p-3 border border-gray-200 rounded bg-gray-50">
              <input type="checkbox" disabled checked className="mt-1" /> 
              <span className="text-gray-700 leading-tight"><strong>Component Deduplication</strong><br/><span className="text-xs text-gray-500">Matches hashes and PURLs.</span></span>
            </label>
            <label className="flex items-start space-x-3 p-3 border border-gray-200 rounded bg-gray-50">
              <input type="checkbox" disabled checked className="mt-1" /> 
              <span className="text-gray-700 leading-tight"><strong>Hierarchy Preservation</strong><br/><span className="text-xs text-gray-500">Maintains structural dependencies.</span></span>
            </label>
            <label className="flex items-start space-x-3 p-3 border border-[#004481] rounded bg-white shadow-sm">
              <input type="checkbox" defaultChecked className="mt-1" /> 
              <span className="text-[#004481] leading-tight"><strong>Conflict Flagging</strong><br/><span className="text-xs text-gray-600">Identifies metadata disagreements.</span></span>
            </label>
          </div>
        </div>

        {/* Footer */}
        <div className="bg-gray-50 -mx-8 -mb-8 p-6 mt-8 border-t border-gray-200 flex justify-between items-center rounded-b-lg">
          <p className="text-xs text-gray-500 italic max-w-sm">
            Merging may take a few seconds depending on file sizes. Redirection will happen automatically.
          </p>
          <button 
            type="submit" 
            disabled={loading || files.length === 0}
            className="bg-[#004481] hover:bg-[#003366] text-white px-8 py-2.5 rounded font-bold shadow transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? "Initializing Aggregation Pipeline..." : "Execute Component Merge"}
          </button>
        </div>
      </form>
    </div>
  );
}
