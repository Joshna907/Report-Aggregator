"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import axios from "axios";
import API_BASE from "@/lib/api";
import { Search, Database, RefreshCw, ChevronDown, Edit3 } from "lucide-react";

interface Component {
  id: number;
  name: string;
  version: string;
  purl: string;
  supplier?: string;
  description?: string;
  licenses?: any[];
  hashes?: any[];
  provenance?: any[];
}

interface MergeResult {
  id: number;
  mergedAt: string;
  summary: any;
  components: Component[];
}

export default function BrowsePage() {
  const [data, setData] = useState<MergeResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [editingComponent, setEditingComponent] = useState<Component | null>(null);
  const [editForm, setEditForm] = useState({ version: "", supplier: "", purl: "", reason: "" });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = () => {
    setLoading(true);
    axios.get(`${API_BASE}/result`)
      .then(res => {
        setData(res.data);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  };

  const handleEditClick = (c: Component) => {
    setEditingComponent(c);
    setEditForm({
      version: c.version || "",
      supplier: c.supplier || "",
      purl: c.purl || "",
      reason: "Manual correction"
    });
  };

  const handleUpdateComponent = async () => {
    if (!editingComponent) return;
    try {
      setLoading(true);
      await axios.post(`${API_BASE}/components/edit`, {
        component: {
          ...editingComponent,
          version: editForm.version,
          supplier: editForm.supplier,
          purl: editForm.purl
        },
        user: "admin",
        reason: editForm.reason
      });
      setEditingComponent(null);
      fetchData();
    } catch (err: any) {
      alert("Update failed: " + (err.response?.data?.error || err.message));
      setLoading(false);
    }
  };

  const handleDownloadJson = () => {
    if (!data) return;
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `aggregator-report-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const handleExportSw360 = async () => {
    try {
      setLoading(true);
      const res = await axios.post(`${API_BASE}/result/export/sw360`);
      alert(res.data.message);
    } catch (err: any) {
      console.error(err);
      alert("Export failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  if (loading && !data) return <div className="p-8 text-sm text-gray-500 flex items-center gap-2"><RefreshCw className="animate-spin text-foss-blue" size={16} /> Loading repository index...</div>;

  return (
    <div className="flex font-sans text-sm min-h-[800px] relative">
      
      {/* Edit Modal */}
      {editingComponent && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-lg overflow-hidden border border-gray-200">
            <div className="bg-gray-50 p-4 border-b border-gray-200 flex justify-between items-center">
              <h3 className="font-bold text-[#004481]">Edit Component: {editingComponent.name}</h3>
              <button onClick={() => setEditingComponent(null)} className="text-gray-400 hover:text-gray-600">&times;</button>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1">
                  <label className="text-xs font-bold text-gray-500 uppercase">Version</label>
                  <input 
                    type="text" 
                    value={editForm.version} 
                    onChange={e => setEditForm({...editForm, version: e.target.value})}
                    className="border border-gray-300 rounded p-2 text-sm focus:border-foss-blue outline-none"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-xs font-bold text-gray-500 uppercase">Supplier</label>
                  <input 
                    type="text" 
                    value={editForm.supplier} 
                    onChange={e => setEditForm({...editForm, supplier: e.target.value})}
                    className="border border-gray-300 rounded p-2 text-sm focus:border-foss-blue outline-none"
                  />
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs font-bold text-gray-500 uppercase">Package URL (PURL)</label>
                <input 
                  type="text" 
                  value={editForm.purl} 
                  onChange={e => setEditForm({...editForm, purl: e.target.value})}
                  className="border border-gray-300 rounded p-2 text-sm font-mono focus:border-foss-blue outline-none"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs font-bold text-gray-500 uppercase">Reason for Change</label>
                <textarea 
                  value={editForm.reason} 
                  onChange={e => setEditForm({...editForm, reason: e.target.value})}
                  className="border border-gray-300 rounded p-2 text-sm h-20 focus:border-foss-blue outline-none"
                />
              </div>
            </div>
            <div className="bg-gray-50 p-4 border-t border-gray-200 flex justify-end gap-3">
              <button 
                onClick={() => setEditingComponent(null)}
                className="px-4 py-2 text-sm font-semibold text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button 
                onClick={handleUpdateComponent}
                className="bg-[#004481] text-white px-6 py-2 rounded text-sm font-bold shadow-sm hover:bg-[#003366] transition-colors"
              >
                Save Changes
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Left Sidebar */}
      <div className="w-64 bg-white border border-gray-200 mr-6 h-fit flex flex-col pb-4 shadow-sm rounded-md overflow-hidden">
        <div className="text-gray-800 font-bold px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center gap-2">
          <Database size={16} className="text-foss-blue" />
          Component Repository
        </div>
        <div className="px-4 mt-4">
          <div className="text-sm font-semibold flex items-center gap-1 cursor-pointer hover:text-foss-blue text-gray-800 mb-2">
            <ChevronDown size={14} /> Local Uploads
          </div>
          <div className="pl-5 mt-1 text-gray-900 bg-blue-50 py-1.5 px-2 rounded font-semibold border border-blue-100 shadow-sm text-xs">
            Aggregated Merges
          </div>
          <Link href="/changelog">
            <div className="pl-5 mt-1 text-gray-500 py-1.5 px-2 text-xs hover:bg-gray-50 cursor-pointer rounded">
              Audit Trail (Changelog)
            </div>
          </Link>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1">
        
        <div className="mb-6 border-b border-gray-200 pb-2">
          <h1 className="text-2xl font-bold text-[#004481] m-0">Repository Browser</h1>
          <p className="text-sm text-gray-500 mt-1">Viewing all fully aggregated, deduplicated components in the system.</p>
        </div>

        {/* Controls bar */}
        <div className="bg-white p-4 rounded-t-md border border-gray-200 border-b-0 flex items-center justify-between shadow-sm">
          <div className="flex items-center space-x-2 text-gray-600">
            <span>Show</span>
            <select className="border border-gray-300 rounded bg-white px-2 py-1 text-gray-800 outline-none focus:border-foss-blue">
              <option>50</option>
              <option>100</option>
            </select>
            <span>entries</span>
          </div>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={14} />
            <input 
              type="text" 
              className="border border-gray-300 rounded-full pl-9 pr-4 py-1.5 bg-gray-50 focus:bg-white focus:border-foss-blue outline-none text-sm w-64" 
              placeholder="Search components or PURLs..." 
            />
          </div>
        </div>

        {/* The Data Table */}
        <div className="overflow-x-auto bg-white border border-gray-200 shadow-sm rounded-b-md">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="bg-gray-50 text-gray-700 border-b border-gray-200">
                <th className="p-3 font-semibold w-[22%]">Component Name & Version</th>
                <th className="p-3 font-semibold text-center">Status</th>
                <th className="p-3 font-semibold w-[22%]">Package URL (PURL)</th>
                <th className="p-3 font-semibold text-center">Main Licenses</th>
                <th className="p-3 font-semibold text-center w-[12%]">Actions</th>
                <th className="p-3 font-semibold text-center">Provenance</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 text-xs">
              {!data || !data.components || data.components.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-8 text-center text-gray-500 bg-gray-50 text-sm">
                    No components found in the repository. Please complete a merge first.
                  </td>
                </tr>
              ) : (
                data.components.map((c, i) => (
                  <tr key={c.id} className="hover:bg-blue-50/50 transition-colors">
                    <td className="p-3 text-[#004481] font-semibold">
                      <div className="hover:underline cursor-pointer inline-block">
                        {c.name} {c.version ? " " + c.version : ""}
                      </div>
                    </td>
                    <td className="p-3 text-center">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-green-100 text-green-800 border border-green-200 uppercase">
                        MERGED
                      </span>
                    </td>
                    <td className="p-3 text-gray-600 font-mono scale-95 origin-left truncate max-w-[180px]" title={c.purl}>
                      {c.purl || "-"}
                    </td>
                    <td className="p-3 text-center">
                      {c.licenses && c.licenses.length > 0 ? (
                        <div className="flex justify-center flex-wrap gap-1">
                          {c.licenses.map(l => (
                            <span key={l.id} className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold bg-gray-100 text-[#c12127] border border-gray-200">
                              {l.id}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-gray-400 italic">NOASSERTION</span>
                      )}
                    </td>
                    <td className="p-3 text-center">
                      <button 
                        onClick={() => handleEditClick(c)}
                        className="text-gray-400 hover:text-[#004481] p-1.5 rounded-full hover:bg-white transition-all shadow-sm flex items-center justify-center mx-auto border border-transparent hover:border-gray-200"
                        title="Edit component metadata"
                      >
                        <Edit3 size={14} />
                      </button>
                    </td>
                    <td className="p-3 text-gray-500 text-center">
                      <div className="flex items-center justify-center gap-1 opacity-70">
                        <Database size={12} />
                        {c.provenance ? c.provenance.length : 1} src
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>

          {/* Table Footer */}
          <div className="bg-gray-50 p-3 border-t border-gray-200 flex justify-between items-center text-xs text-gray-500">
            <span>Showing 1 to {data?.components?.length || 0} of {data?.components?.length || 0} entries</span>
            
            <div className="flex gap-2">
              <button 
                onClick={handleExportSw360}
                className="bg-white border border-gray-300 hover:bg-gray-100 px-3 py-1 rounded shadow-sm text-gray-700 font-medium transition-colors disabled:opacity-50"
                disabled={!data || loading}
              >
                Export to SW360
              </button>
              <button 
                onClick={handleDownloadJson}
                className="bg-white border border-gray-300 hover:bg-gray-100 px-3 py-1 rounded shadow-sm text-gray-700 font-medium transition-colors disabled:opacity-50"
                disabled={!data}
              >
                Download JSON
              </button>
            </div>
          </div>

        </div>

      </div>
    </div>
  );
}
