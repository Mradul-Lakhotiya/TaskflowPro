"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/authStore";

export default function Providers({ children }: { children: React.ReactNode }) {
  const initializeAuth = useAuthStore((state) => state.initialize);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    initializeAuth();
    setMounted(true);
  }, [initializeAuth]);

  // Prevent hydration mismatch by only rendering children when mounted
  if (!mounted) {
    return null;
  }

  return <>{children}</>;
}
