"use client";

import { useState, useEffect } from "react";
import { GetVersion } from "@/wailsjs/go/main/App";

export function Footer() {
  const [appVersion, setAppVersion] = useState("...");

  useEffect(() => {
    GetVersion().then(setAppVersion).catch(() => setAppVersion("unknown"));
  }, []);

  return (
    <footer className="border-t border-border/40 bg-background/40 backdrop-blur-sm">
      <div className="max-w-5xl mx-auto px-4 py-2 flex justify-between items-center text-xs text-muted-foreground">
        <span>© Y3L Corp</span>
        <span>{appVersion}</span>
      </div>
    </footer>
  );
}
