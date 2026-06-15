"use client";

import { useEffect, useState, useCallback } from "react";
import { Plus, Search, Filter, LogOut, ChevronLeft, ChevronRight, Users } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import ProtectedRoute from "@/components/ProtectedRoute";
import TaskCard, { Task } from "@/components/TaskCard";
import TaskFormModal from "@/components/TaskFormModal";
import CustomSelect from "@/components/CustomSelect";
import { useAuthStore } from "@/store/authStore";
import api from "@/lib/api";

export default function Dashboard() {
  const { user, token, logout } = useAuthStore();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  
  // Admin User Filter
  const [users, setUsers] = useState<{id: number, email: string}[]>([]);
  const [selectedUserId, setSelectedUserId] = useState<string>("");
  
  // Filters & Sorting & Pagination
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [sortBy, setSortBy] = useState("created_at");
  const [sortDesc, setSortDesc] = useState(true);
  const [page, setPage] = useState(1);
  const [totalTasks, setTotalTasks] = useState(0);
  const limit = 10;

  // Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [taskToEdit, setTaskToEdit] = useState<Task | null>(null);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await api.get("/tasks", {
        params: {
          page,
          limit,
          search,
          status,
          sort_by: sortBy,
          sort_desc: sortDesc,
          user_id: selectedUserId || undefined,
        },
      });
      setTasks(data.data || []);
      setTotalTasks(data.total || 0);
    } catch (error) {
      console.error("Failed to fetch tasks", error);
    } finally {
      setLoading(false);
    }
  }, [page, limit, search, status, sortBy, sortDesc, selectedUserId]);

  // Debounced search
  useEffect(() => {
    const handler = setTimeout(() => {
      setPage(1); // Reset page on new search
      fetchTasks();
    }, 300);
    return () => clearTimeout(handler);
  }, [search, status, sortBy, sortDesc, selectedUserId, fetchTasks]);

  // Fetch users for admin filter
  useEffect(() => {
    if (user?.role === 'admin') {
      api.get("/users")
        .then(res => setUsers(res.data || []))
        .catch(err => console.error("Failed to fetch users", err));
    }
  }, [user]);

  // Real-time SSE connection
  useEffect(() => {
    if (!token) return;

    // Use Next.js API rewrite path (vital for Codespaces so browser doesn't try to hit localhost directly)
    const url = `/api/events?token=${token}`;
    const eventSource = new EventSource(url);

    eventSource.onopen = () => {
      console.log("SSE connected");
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("SSE Event Received:", data);
        
        // Skip the initial connected event
        if (data.status === "connected") return;

        // On any task event, re-fetch tasks to guarantee exact sorting/pagination
        fetchTasks();
      } catch (err) {
        console.error("Failed to parse SSE event", err);
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE Error:", error);
      // Let the browser automatically reconnect
    };

    return () => {
      eventSource.close();
    };
  }, [token, fetchTasks]);

  const handleUpdateTask = (updatedTask: Task) => {
    setTasks(prev => prev.map(t => t.id === updatedTask.id ? updatedTask : t));
  };

  const handleDeleteTask = (id: number) => {
    setTasks(prev => prev.filter(t => t.id !== id));
    setTotalTasks(prev => prev - 1);
  };

  const handleModalSuccess = (task: Task, isEdit: boolean) => {
    if (isEdit) {
      handleUpdateTask(task);
    } else {
      fetchTasks(); // Refetch to apply sorting/pagination correctly for new item
    }
  };

  const openEditModal = (task: Task) => {
    setTaskToEdit(task);
    setIsModalOpen(true);
  };

  const totalPages = Math.max(1, Math.ceil(totalTasks / limit));

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-background text-foreground pb-20">
        {/* Header */}
        <header className="sticky top-0 z-40 bg-background/80 backdrop-blur-md border-b border-border shadow-sm">
          <div className="max-w-5xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-bold shadow-lg shadow-primary/20">
                R
              </div>
              <h1 className="font-semibold text-lg hidden sm:block">Rival Tasks</h1>
            </div>
            
            <div className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground hidden sm:block">
                {user?.email} {user?.role === 'admin' && <span className="ml-2 px-2 py-0.5 bg-accent rounded text-xs">Admin</span>}
              </span>
              <button 
                onClick={logout}
                className="p-2 text-muted-foreground hover:text-destructive transition-colors rounded-full hover:bg-muted"
                title="Logout"
              >
                <LogOut size={18} />
              </button>
            </div>
          </div>
        </header>

        <main className="max-w-5xl mx-auto px-4 sm:px-6 pt-8">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
            <h2 className="text-3xl font-bold tracking-tight">Your Tasks</h2>
            <button
              onClick={() => {
                setTaskToEdit(null);
                setIsModalOpen(true);
              }}
              className="bg-primary text-primary-foreground px-5 py-2.5 rounded-full font-medium shadow-lg shadow-primary/25 hover:shadow-xl hover:shadow-primary/30 hover:-translate-y-0.5 transition-all flex items-center gap-2 w-full sm:w-auto justify-center"
            >
              <Plus size={20} />
              New Task
            </button>
          </div>

          {/* Controls Bar */}
          <div className="glass p-3 rounded-[var(--radius-lg)] mb-8 flex flex-col md:flex-row gap-3 shadow-sm border border-border/50">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" size={18} />
              <input
                type="text"
                placeholder="Search tasks..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full bg-input/50 border border-transparent hover:border-border focus:border-border rounded-[var(--radius-md)] pl-10 pr-4 py-2.5 outline-none transition-all"
              />
            </div>
            
            <div className="flex gap-3 flex-wrap">
              {user?.role === 'admin' && (
                <CustomSelect
                  value={selectedUserId}
                  onChange={(val) => setSelectedUserId(val)}
                  icon={<Users size={16} />}
                  className="w-full sm:w-[160px]"
                  placeholder="All Users"
                  options={[
                    { value: "", label: "All Users" },
                    ...users.map(u => ({ value: u.id.toString(), label: u.email }))
                  ]}
                />
              )}

              <CustomSelect
                value={status}
                onChange={(val) => setStatus(val)}
                icon={<Filter size={16} />}
                className="w-full sm:w-[160px]"
                placeholder="All Status"
                options={[
                  { value: "", label: "All Status" },
                  { value: "pending", label: "Pending" },
                  { value: "in_progress", label: "In Progress" },
                  { value: "completed", label: "Completed" },
                ]}
              />

              <CustomSelect
                value={`${sortBy}-${sortDesc}`}
                onChange={(val) => {
                  const [sBy, sDesc] = val.split('-');
                  setSortBy(sBy);
                  setSortDesc(sDesc === 'true');
                  setPage(1);
                }}
                className="w-full sm:w-[220px]"
                options={[
                  { value: "created_at-true", label: "Newest First" },
                  { value: "created_at-false", label: "Oldest First" },
                  { value: "due_date-false", label: "Due Date (Soonest)" },
                  { value: "due_date-true", label: "Due Date (Latest)" },
                  { value: "priority-false", label: "Priority (High to Low)" },
                  { value: "priority-true", label: "Priority (Low to High)" },
                ]}
              />
            </div>
          </div>

          {/* Task Grid */}
          {loading && tasks.length === 0 ? (
            <div className="py-20 flex justify-center">
              <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin" />
            </div>
          ) : tasks.length === 0 ? (
            <motion.div 
              initial={{ opacity: 0 }} 
              animate={{ opacity: 1 }}
              className="py-20 text-center glass rounded-[var(--radius-lg)] border border-dashed border-border"
            >
              <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4 text-muted-foreground">
                <Search size={24} />
              </div>
              <h3 className="text-xl font-medium mb-2">No tasks found</h3>
              <p className="text-muted-foreground max-w-md mx-auto">
                {search || status ? "Try adjusting your filters or search query." : "You're all caught up! Create a new task to get started."}
              </p>
            </motion.div>
          ) : (
            <div className="space-y-4">
              <AnimatePresence mode="popLayout">
                {tasks.map(task => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    onUpdate={handleUpdateTask}
                    onDelete={handleDeleteTask}
                    onEdit={openEditModal}
                  />
                ))}
              </AnimatePresence>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-10 flex items-center justify-center gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-2 rounded-full hover:bg-muted disabled:opacity-50 transition-colors"
              >
                <ChevronLeft size={20} />
              </button>
              
              <span className="text-sm font-medium px-4">
                Page {page} of {totalPages}
              </span>

              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="p-2 rounded-full hover:bg-muted disabled:opacity-50 transition-colors"
              >
                <ChevronRight size={20} />
              </button>
            </div>
          )}
        </main>
      </div>

      <TaskFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        taskToEdit={taskToEdit}
        onSuccess={handleModalSuccess}
      />
    </ProtectedRoute>
  );
}
