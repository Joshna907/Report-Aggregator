"use client";

import React, { useEffect, useState } from "react";
import axios from "axios";
import API_BASE from "@/lib/api";
import { AlertTriangle, CheckCircle2, ListChecks, ChevronDown } from "lucide-react";

export default function ConflictsPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [resolving, setResolving] = useState<number | null>(null);

  const fetchResult = () => {
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

  useEffect(() => {
    fetchResult();
  }, []);

  const handleResolve = async (conflict: any, chosenValue: string) => {
    setResolving(conflict.id);
    try {
      await axios.post(`${API_BASE}/conflicts/resolve`, {
        id: conflict.id,
        reportId: conflict.reportId,
        componentName: conflict.componentName,
        componentVer: conflict.componentVer,
        field: conflict.field,
        resolution: chosenValue,
      });
      fetchResult();
    } catch (err) {
      console.error(err);
      alert("Failed to resolve conflict.");
    } finally {
      setResolving(null);
    }
  };

  if (loading && !data) return <div className="p-8 text-sm text-gray-500">Loading conflicts...</div>;

  const conflicts = data?.conflicts || [];
  const unresolved = conflicts.filter((c: any) => !c.resolved);
  const resolved = conflicts.filter((c: any) => c.resolved);

  return (
    <div className="flex font-sans text-sm min-h-[800px]">
      
      {/* Left Sidebar */}
      <div className="w-64 bg-white border border-gray-200 mr-6 h-fit flex flex-col pb-4 shadow-sm rounded-md overflow-hidden">
        <div className="text-gray-800 font-bold px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center gap-2">
          <ListChecks size={16} className="text-[#004481]" />
          Organize Tasks
        </div>
        <div className="px-4 mt-4">
          <div className="text-sm font-semibold flex items-center gap-1 cursor-pointer hover:text-foss-blue text-gray-800 mb-2">
            <ChevronDown size={14} /> Active Uploads
          </div>
          <div className="pl-5 mt-1 text-red-800 bg-red-50 py-1.5 px-2 rounded font-semibold border border-red-100 shadow-sm text-xs flex justify-between items-center">
            <span>Resolve Conflicts</span>
            <span className="bg-red-200 text-red-800 px-1.5 rounded-full text-[10px]">{unresolved.length}</span>
          </div>
          <div className="pl-5 mt-1 text-gray-600 py-1.5 px-2 text-xs hover:bg-gray-50 cursor-pointer rounded flex justify-between items-center transition-colors">
            <span>Cleared Items</span>
            <span className="bg-gray-200 text-gray-700 px-1.5 rounded-full text-[10px]">{resolved.length}</span>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1">
        <div className="mb-6 border-b border-gray-200 pb-2">
          <h1 className="text-2xl font-bold text-[#004481] m-0 flex items-center gap-2">
            Metadata Conflict Resolution
          </h1>
          <p className="text-sm text-gray-500 mt-1 max-w-3xl">
            The aggregator found disagreements between the SBOMs you merged. Please review the conflicting fields below and select the correct value to establish a unified source of truth.
          </p>
        </div>

        {unresolved.length === 0 ? (
          <div className="p-6 bg-green-50 border border-green-200 text-green-800 font-bold mb-4 w-full rounded-lg shadow-sm flex items-center gap-3">
            <CheckCircle2 size={24} className="text-green-600" />
            <div>
              <div className="text-lg">No active conflicts to resolve.</div>
              <div className="text-sm font-normal text-green-700 mt-1">All aggregated data is perfectly synchronized!</div>
            </div>
          </div>
        ) : (
          <div className="overflow-hidden bg-white border border-gray-200 shadow-sm rounded-lg">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="bg-gray-50 text-gray-700 border-b border-gray-200 text-xs uppercase tracking-wider">
                  <th className="p-4 font-semibold">Component</th>
                  <th className="p-4 font-semibold text-center">Disputed Field</th>
                  <th className="p-4 font-semibold text-center w-[30%] border-l border-gray-200">Option A</th>
                  <th className="p-4 font-semibold text-center w-[30%] pl-0">Option B</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {unresolved.map((c: any, i: number) => (
                  <tr key={c.id} className="hover:bg-gray-50 transition-colors">
                    <td className="p-4 font-bold text-gray-800">
                      {c.componentName} <span className="font-normal text-gray-500 block text-xs mt-1">v{c.componentVer}</span>
                    </td>
                    <td className="p-4 text-center">
                      <span className="inline-flex items-center gap-1 bg-red-50 text-[#c12127] border border-red-100 px-2 py-1 rounded text-xs font-bold uppercase tracking-wider">
                        <AlertTriangle size={12} /> {c.field}
                      </span>
                    </td>
                    
                    {/* Option A */}
                    <td className="p-4 text-center border-l border-gray-200 bg-white group hover:bg-red-50/30 transition-colors">
                      <div className="text-gray-900 font-semibold mb-2 bg-gray-50 p-2 rounded border border-gray-100 min-h-[40px] flex items-center justify-center">
                        {c.valueA}
                      </div>
                      <div className="text-[11px] text-gray-500 italic mb-3">Source: {c.sourceA}</div>
                      <button 
                        onClick={() => handleResolve(c, c.valueA)}
                        disabled={resolving === c.id}
                        className="w-full bg-white border-2 border-gray-300 hover:border-[#004481] hover:text-[#004481] text-gray-700 py-1.5 rounded text-xs font-bold transition-all disabled:opacity-50"
                      >
                        Keep Option A
                      </button>
                    </td>

                    {/* Option B */}
                    <td className="p-4 text-center bg-white group hover:bg-blue-50/30 transition-colors">
                      <div className="text-gray-900 font-semibold mb-2 bg-gray-50 p-2 rounded border border-gray-100 min-h-[40px] flex items-center justify-center">
                        {c.valueB}
                      </div>
                      <div className="text-[11px] text-gray-500 italic mb-3">Source: {c.sourceB}</div>
                      <button 
                        onClick={() => handleResolve(c, c.valueB)}
                        disabled={resolving === c.id}
                        className="w-full bg-white border-2 border-gray-300 hover:border-[#004481] hover:text-[#004481] text-gray-700 py-1.5 rounded text-xs font-bold transition-all disabled:opacity-50"
                      >
                        Keep Option B
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

      </div>
    </div>
  );
}
