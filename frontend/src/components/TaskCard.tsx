"use client";

import { useState } from "react";
import { format } from "date-fns";
import { motion } from "framer-motion";
import { CheckCircle2, Circle, Clock, Trash2, Edit2, AlertCircle, Paperclip, History } from "lucide-react";
import api from "@/lib/api";

export interface TaskAttachment {
  id: number;
  task_id: number;
  file_name: string;
  file_url: string;
  created_at: string;
}

export interface Task {
  id: number;
  title: string;
  description: string;
  status: "pending" | "in_progress" | "completed";
  priority: "low" | "medium" | "high";
  attachments: TaskAttachment[];
  due_date: string | null;
  created_at: string;
  updated_at: string;
}

interface TaskCardProps {
  task: Task;
  onUpdate: (task: Task) => void;
  onDelete: (id: number) => void;
  onEdit: (task: Task) => void;
  onActivityClick: (taskId: number) => void;
}

const priorityColors = {
  high: "text-destructive bg-destructive/10 border-destructive/20",
  medium: "text-yellow-600 bg-yellow-500/10 border-yellow-500/20",
  low: "text-green-600 bg-green-500/10 border-green-500/20",
};

export default function TaskCard({ task, onUpdate, onDelete, onEdit, onActivityClick }: TaskCardProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isToggling, setIsToggling] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const toggleStatus = async () => {
    if (isToggling) return;
    setIsToggling(true);
    const newStatus = task.status === "completed" ? "pending" : "completed";
    
    const previousStatus = task.status;
    onUpdate({ ...task, status: newStatus });

    try {
      await api.patch(`/tasks/${task.id}`, { status: newStatus });
    } catch (error) {
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

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;

    setIsUploading(true);
    try {
      let latestTask = task;
      await Promise.all(
        Array.from(files).map(async (file) => {
          const formData = new FormData();
          formData.append('file', file);
          const response = await api.post(`/tasks/${task.id}/upload`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
          latestTask = response.data;
        })
      );
      onUpdate(latestTask);
    } catch (err) {
      console.error("Failed to upload file(s)", err);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDeleteAttachment = async (attachmentId: number) => {
    try {
      const response = await api.delete(`/tasks/${task.id}/attachments/${attachmentId}`);
      onUpdate(response.data);
    } catch (err) {
      console.error("Failed to delete attachment", err);
    }
  };

  const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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
            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <input 
                type="file" 
                multiple
                id={`upload-${task.id}`}
                className="hidden" 
                onChange={handleFileUpload}
                disabled={isUploading}
              />
              <label 
                htmlFor={`upload-${task.id}`}
                className="p-1.5 text-muted-foreground hover:text-primary bg-background/50 rounded-md transition-colors cursor-pointer"
                title="Upload Attachment"
              >
                <Paperclip size={16} />
              </label>
              <button 
                onClick={() => onActivityClick(task.id)} 
                className="p-1.5 text-muted-foreground hover:text-primary bg-background/50 rounded-md transition-colors"
                title="View Activity Log"
              >
                <History size={16} />
              </button>
              <button 
                onClick={() => onEdit(task)} 
                className="p-1.5 text-muted-foreground hover:text-primary bg-background/50 rounded-md transition-colors"
                title="Edit Task"
              >
                <Edit2 size={16} />
              </button>
              <button onClick={handleDelete} disabled={isDeleting} className="p-1.5 text-muted-foreground hover:text-destructive bg-background/50 rounded-md transition-colors">
                <Trash2 size={16} />
              </button>
            </div>
          </div>
          
          {task.description && (
            <p className="text-sm text-muted-foreground mb-3 line-clamp-2">{task.description}</p>
          )}

          {task.attachments && task.attachments.length > 0 && (
            <div className="mt-2 mb-3 flex flex-col gap-2">
              {task.attachments.map((att) => {
                const isImg = att.file_url.match(/\.(jpeg|jpg|gif|png|webp)$/i);
                const fullUrl = `${backendUrl}${att.file_url}`;
                return (
                  <div key={att.id} className="relative group/att border rounded-lg overflow-hidden border-border bg-muted/30">
                    <button
                      onClick={() => handleDeleteAttachment(att.id)}
                      className="absolute top-1 right-1 p-1 bg-background/80 hover:bg-destructive hover:text-destructive-foreground rounded text-muted-foreground opacity-0 group-hover/att:opacity-100 transition-all z-10"
                      title="Remove Attachment"
                    >
                      <Trash2 size={12} />
                    </button>
                    {isImg ? (
                      <div className="relative">
                        <img src={fullUrl} alt={att.file_name} className="w-full h-auto object-cover max-h-32" />
                        <div className="absolute bottom-0 left-0 right-0 bg-background/80 p-1 text-[10px] truncate px-2 text-muted-foreground">{att.file_name}</div>
                      </div>
                    ) : (
                      <a href={fullUrl} target="_blank" rel="noopener noreferrer" className="flex items-center p-2 text-xs text-primary hover:underline">
                        <Paperclip size={12} className="mr-2 flex-shrink-0" /> 
                        <span className="truncate">{att.file_name}</span>
                      </a>
                    )}
                  </div>
                );
              })}
            </div>
          )}

          <div className="flex flex-wrap items-center gap-3 text-xs font-medium">
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
