"use client";

import { useState } from "react";
import { format } from "date-fns";
import { motion } from "framer-motion";
import { CheckCircle2, Circle, Clock, Trash2, Edit2, AlertCircle } from "lucide-react";
import api from "@/lib/api";

export interface Task {
  id: number;
  title: string;
  description: string;
  status: "pending" | "in_progress" | "completed";
  priority: "low" | "medium" | "high";
  due_date: string | null;
  created_at: string;
}

interface TaskCardProps {
  task: Task;
  onUpdate: (task: Task) => void;
  onDelete: (id: number) => void;
  onEdit: (task: Task) => void;
}

const priorityColors = {
  high: "text-destructive bg-destructive/10 border-destructive/20",
  medium: "text-yellow-600 bg-yellow-500/10 border-yellow-500/20",
  low: "text-green-600 bg-green-500/10 border-green-500/20",
};

export default function TaskCard({ task, onUpdate, onDelete, onEdit }: TaskCardProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isToggling, setIsToggling] = useState(false);

  const toggleStatus = async () => {
    if (isToggling) return;
    setIsToggling(true);
    const newStatus = task.status === "completed" ? "pending" : "completed";
    
    // Optimistic UI update
    const previousStatus = task.status;
    onUpdate({ ...task, status: newStatus });

    try {
      await api.patch(`/tasks/${task.id}`, { status: newStatus });
    } catch (error) {
      // Rollback on failure
      onUpdate({ ...task, status: previousStatus });
      console.error("Failed to update status", error);
    } finally {
      setIsToggling(false);
    }
  };

  const handleDelete = async () => {
    if (isDeleting) return;
    setIsDeleting(true);
    try {
      await api.delete(`/tasks/${task.id}`);
      onDelete(task.id);
    } catch (error) {
      console.error("Failed to delete task", error);
      setIsDeleting(false);
    }
  };

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      whileHover={{ y: -2 }}
      className={`glass p-5 rounded-[var(--radius-lg)] shadow-sm transition-all relative overflow-hidden group ${
        task.status === "completed" ? "opacity-75" : ""
      }`}
    >
      <div className="flex items-start gap-4">
        <button
          onClick={toggleStatus}
          disabled={isToggling}
          className="mt-1 flex-shrink-0 text-muted-foreground hover:text-primary transition-colors focus:outline-none"
        >
          {task.status === "completed" ? (
            <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} className="text-primary">
              <CheckCircle2 size={24} />
            </motion.div>
          ) : (
            <Circle size={24} />
          )}
        </button>

        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between gap-2 mb-1">
            <h3 className={`font-semibold text-lg truncate ${task.status === "completed" ? "line-through text-muted-foreground" : ""}`}>
              {task.title}
            </h3>
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button onClick={() => onEdit(task)} className="p-1.5 text-muted-foreground hover:text-primary bg-background/50 rounded-md transition-colors">
                <Edit2 size={16} />
              </button>
              <button onClick={handleDelete} disabled={isDeleting} className="p-1.5 text-muted-foreground hover:text-destructive bg-background/50 rounded-md transition-colors">
                <Trash2 size={16} />
              </button>
            </div>
          </div>
          
          {task.description && (
            <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{task.description}</p>
          )}

          <div className="flex flex-wrap items-center gap-3 text-xs font-medium mt-3">
            <span className={`px-2.5 py-1 rounded-full border capitalize flex items-center gap-1 ${priorityColors[task.priority]}`}>
              {task.priority === 'high' && <AlertCircle size={12} />}
              {task.priority} Priority
            </span>
            
            {task.due_date && (
              <span className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full border ${
                new Date(task.due_date) < new Date() && task.status !== "completed" 
                  ? "text-destructive border-destructive/30 bg-destructive/5" 
                  : "text-muted-foreground border-border bg-muted/30"
              }`}>
                <Clock size={12} />
                {format(new Date(task.due_date), "MMM d, yyyy")}
              </span>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
}
