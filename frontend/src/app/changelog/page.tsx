"use client";

import React, { useEffect, useState } from "react";
import axios from "axios";
import API_BASE from "@/lib/api";
import { History, User, Clock, FileText, ChevronRight } from "lucide-react";

interface AuditLog {
  id: number;
  componentName: string;
  field: string;
  oldValue: string;
  newValue: string;
  changedBy: string;
  reason: string;
  timestamp: string;
}

export default function ChangelogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // We fetch the latest report result first to get the report ID
    axios.get(`${API_BASE}/result`)
      .then(res => {
        if (res.data && res.data.id) {
          return axios.get(`${API_BASE}/changelog/${res.data.id}`);
        }
        return { data: [] };
      })
      .then(res => {
        setLogs(res.data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  if (loading) return <div className="p-8 text-sm text-gray-500">Loading audit trail...</div>;

  return (
    <div className="max-w-6xl mx-auto py-8 px-4 font-sans text-gray-800 animate-in fade-in duration-500">
      <div className="mb-8 border-b border-gray-200 pb-6 flex justify-between items-end">
        <div>
          <h1 className="text-3xl font-bold text-[#004481] flex items-center gap-3">
            <History size={32} />
            System Audit Trail
          </h1>
          <p className="text-sm text-gray-500 mt-2">
            Complete record of manual metadata overrides and automated reconciliation decisions.
          </p>
        </div>
        <div className="bg-gray-100 px-4 py-2 rounded-md text-xs font-mono text-gray-600 border border-gray-200 shadow-sm">
          Total Entries: {logs.length}
        </div>
      </div>

      {logs.length === 0 ? (
        <div className="bg-white border border-dashed border-gray-300 rounded-xl p-16 text-center">
          <div className="bg-gray-50 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 text-gray-300">
            <Clock size={32} />
          </div>
          <h3 className="text-lg font-semibold text-gray-600">No audit logs recorded yet</h3>
          <p className="text-sm text-gray-400 mt-1">Manual edits to components will appear here in real-time.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {logs.map((log) => (
            <div key={log.id} className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden group hover:border-[#004481] transition-all">
              <div className="p-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-start gap-4 flex-1">
                  <div className="bg-blue-50 p-3 rounded-lg text-[#004481]">
                    <FileText size={20} />
                  </div>
                  <div>
                    <h3 className="font-bold text-gray-900 flex items-center gap-2">
                      {log.componentName}
                      <span className="text-[10px] bg-gray-100 text-gray-500 px-2 py-0.5 rounded border border-gray-200 uppercase tracking-tighter">
                        {log.field}
                      </span>
                    </h3>
                    <div className="flex items-center gap-4 mt-1 text-xs text-gray-500">
                      <div className="flex items-center gap-1.5">
                        <User size={12} /> {log.changedBy}
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Clock size={12} /> {new Date(log.timestamp).toLocaleString()}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="flex-1 flex items-center gap-3 bg-gray-50 p-3 rounded-md border border-gray-100">
                  <div className="flex-1 text-center min-w-0">
                    <p className="text-[10px] uppercase font-bold text-gray-400 mb-1">Old Value</p>
                    <p className="text-xs text-red-600 truncate font-mono bg-white px-2 py-1 rounded border border-red-50">{log.oldValue || "(Empty)"}</p>
                  </div>
                  <ChevronRight size={16} className="text-gray-300" />
                  <div className="flex-1 text-center min-w-0">
                    <p className="text-[10px] uppercase font-bold text-gray-400 mb-1">New Value</p>
                    <p className="text-xs text-green-700 font-bold truncate font-mono bg-white px-2 py-1 rounded border border-green-50">{log.newValue}</p>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50/50 px-4 py-2 border-t border-gray-100 text-[11px] text-gray-500 flex items-center gap-2">
                <span className="font-bold uppercase tracking-widest text-[#004481] opacity-70">Reason:</span>
                {log.reason}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
