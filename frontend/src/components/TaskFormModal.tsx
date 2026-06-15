"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Loader2 } from "lucide-react";
import api from "@/lib/api";
import { Task } from "./TaskCard";

interface TaskFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (task: Task, isEdit: boolean) => void;
  taskToEdit?: Task | null;
}

export default function TaskFormModal({ isOpen, onClose, onSuccess, taskToEdit }: TaskFormModalProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<"low" | "medium" | "high">("medium");
  const [status, setStatus] = useState<"pending" | "in_progress" | "completed">("pending");
  const [dueDate, setDueDate] = useState("");
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (taskToEdit) {
      setTitle(taskToEdit.title);
      setDescription(taskToEdit.description || "");
      setPriority(taskToEdit.priority);
      setStatus(taskToEdit.status);
      setDueDate(taskToEdit.due_date ? new Date(taskToEdit.due_date).toISOString().split('T')[0] : "");
    } else {
      setTitle("");
      setDescription("");
      setPriority("medium");
      setStatus("pending");
      setDueDate("");
    }
  }, [taskToEdit, isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError("Title is required");
      return;
    }
    
    setLoading(true);
    setError("");

    const payload = {
      title,
      description,
      priority,
      status,
      due_date: dueDate ? new Date(dueDate).toISOString() : null,
    };

    try {
      if (taskToEdit) {
        const { data } = await api.patch(`/tasks/${taskToEdit.id}`, payload);
        onSuccess(data, true);
      } else {
        const { data } = await api.post("/tasks", payload);
        onSuccess(data, false);
      }
      onClose();
    } catch (err: any) {
      setError(err.response?.data || "An error occurred while saving the task.");
    } finally {
      setLoading(false);
    }
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
              className="w-full max-w-lg bg-background border border-border shadow-2xl rounded-[var(--radius-lg)] overflow-hidden pointer-events-auto"
            >
              <div className="flex items-center justify-between p-6 border-b border-border/50 bg-muted/20">
                <h2 className="text-xl font-semibold tracking-tight">
                  {taskToEdit ? "Edit Task" : "Create New Task"}
                </h2>
                <button
                  onClick={onClose}
                  className="p-2 hover:bg-muted rounded-full transition-colors text-muted-foreground"
                >
                  <X size={20} />
                </button>
              </div>

              <form onSubmit={handleSubmit} className="p-6 space-y-5">
                {error && (
                  <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-[var(--radius-md)]">
                    {error}
                  </div>
                )}

                <div>
                  <label className="block text-sm font-medium mb-1.5">Task Title *</label>
                  <input
                    type="text"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    className="w-full bg-input/50 border border-border rounded-[var(--radius-md)] px-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-ring"
                    placeholder="e.g. Complete quarterly report"
                    autoFocus
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1.5">Description</label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={3}
                    className="w-full bg-input/50 border border-border rounded-[var(--radius-md)] px-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                    placeholder="Add details about this task..."
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
                  <div>
                    <label className="block text-sm font-medium mb-1.5">Priority</label>
                    <select
                      value={priority}
                      onChange={(e) => setPriority(e.target.value as any)}
                      className="w-full bg-input/50 border border-border rounded-[var(--radius-md)] px-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-ring"
                    >
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1.5">Due Date</label>
                    <input
                      type="date"
                      value={dueDate}
                      onChange={(e) => setDueDate(e.target.value)}
                      className="w-full bg-input/50 border border-border rounded-[var(--radius-md)] px-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                  </div>
                </div>

                {taskToEdit && (
                  <div>
                    <label className="block text-sm font-medium mb-1.5">Status</label>
                    <select
                      value={status}
                      onChange={(e) => setStatus(e.target.value as any)}
                      className="w-full bg-input/50 border border-border rounded-[var(--radius-md)] px-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-ring"
                    >
                      <option value="pending">Pending</option>
                      <option value="in_progress">In Progress</option>
                      <option value="completed">Completed</option>
                    </select>
                  </div>
                )}

                <div className="pt-4 border-t border-border/50 flex items-center justify-end gap-3 mt-6">
                  <button
                    type="button"
                    onClick={onClose}
                    className="px-4 py-2 rounded-[var(--radius-md)] font-medium hover:bg-muted transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={loading}
                    className="px-6 py-2 bg-primary text-primary-foreground rounded-[var(--radius-md)] font-medium flex items-center gap-2 hover:bg-primary/90 transition-colors disabled:opacity-70"
                  >
                    {loading && <Loader2 size={16} className="animate-spin" />}
                    {taskToEdit ? "Save Changes" : "Create Task"}
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  );
}
