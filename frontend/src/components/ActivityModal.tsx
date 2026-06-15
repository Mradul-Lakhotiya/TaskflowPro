/* eslint-disable react-hooks/set-state-in-effect */
/* eslint-disable @typescript-eslint/no-unused-vars */
"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Loader2, History, CheckCircle2, Edit2, FileText, PlusCircle } from "lucide-react";
import { format } from "date-fns";
import api from "@/lib/api";

interface Activity {
  id: number;
  task_id: number;
  user_email: string;
  action: string;
  details: string | null;
  created_at: string;
}

interface ActivityModalProps {
  isOpen: boolean;
  onClose: () => void;
  taskId: number | null;
}

export default function ActivityModal({ isOpen, onClose, taskId }: ActivityModalProps) {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isOpen && taskId) {
      setLoading(true);
      api.get(`/tasks/${taskId}/activity`)
        .then((res) => setActivities(res.data || []))
        .catch((err) => setError("Failed to load activity log."))
        .finally(() => setLoading(false));
    } else {
      setActivities([]);
      setError("");
    }
  }, [isOpen, taskId]);

  const getActionIcon = (action: string) => {
    switch (action) {
      case "created": return <PlusCircle size={16} className="text-green-500" />;
      case "updated": return <Edit2 size={16} className="text-blue-500" />;
      case "attachment_uploaded": return <FileText size={16} className="text-purple-500" />;
      case "status_changed": return <CheckCircle2 size={16} className="text-emerald-500" />;
      default: return <History size={16} className="text-gray-500" />;
    }
  };

  const getActionText = (action: string) => {
    return action.replace(/_/g, " ");
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50"
          />
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-0 pointer-events-none">
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              className="w-full max-w-lg max-h-[85vh] flex flex-col bg-background border border-border shadow-2xl rounded-[var(--radius-lg)] overflow-hidden pointer-events-auto"
            >
              <div className="flex items-center justify-between p-6 border-b border-border/50 bg-muted/20">
                <div className="flex items-center gap-2">
                  <History size={20} className="text-muted-foreground" />
                  <h2 className="text-xl font-semibold tracking-tight">Activity Log</h2>
                </div>
                <button
                  onClick={onClose}
                  className="p-2 hover:bg-muted rounded-full transition-colors text-muted-foreground"
                >
                  <X size={20} />
                </button>
              </div>

              <div className="p-6 overflow-y-auto no-scrollbar flex-1">
                {error && (
                  <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-[var(--radius-md)] mb-4">
                    {error}
                  </div>
                )}

                {loading ? (
                  <div className="flex justify-center py-10">
                    <Loader2 className="animate-spin text-primary" size={24} />
                  </div>
                ) : activities.length === 0 ? (
                  <div className="text-center py-10 text-muted-foreground">
                    No activity recorded for this task.
                  </div>
                ) : (
                  <div className="relative pl-4 border-l-2 border-muted space-y-6">
                    {activities.map((activity, index) => (
                      <motion.div 
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        key={activity.id} 
                        className="relative"
                      >
                        <div className="absolute -left-[25px] bg-background p-1 rounded-full border border-border">
                          {getActionIcon(activity.action)}
                        </div>
                        <div className="glass p-3 rounded-lg border border-border/50 text-sm">
                          <div className="flex justify-between items-start mb-1">
                            <span className="font-medium capitalize">{getActionText(activity.action)}</span>
                            <span className="text-xs text-muted-foreground whitespace-nowrap ml-2">
                              {format(new Date(activity.created_at), "MMM d, h:mm a")}
                            </span>
                          </div>
                          {activity.details && (
                            <p className="text-muted-foreground mt-1 bg-muted/30 p-2 rounded text-xs break-words">
                              {activity.details}
                            </p>
                          )}
                          <div className="text-xs text-muted-foreground mt-2 text-right opacity-70">
                            by {activity.user_email || "Unknown"}
                          </div>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </div>
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  );
}
