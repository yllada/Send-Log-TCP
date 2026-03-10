"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "@/hooks/use-toast";
import {
  SelectLogFile,
  ImportLogFile,
} from "@/wailsjs/go/main/BatchImportService";
import {
  FileUpIcon,
  FileText,
  AlertCircleIcon,
  CheckCircle2Icon,
  Loader2Icon,
} from "lucide-react";

interface BatchImportResult {
  messages: string[];
  totalLines: number;
  errors: string[];
}

interface BatchImportProps {
  onImport: (messages: string[]) => void;
  currentMessages?: string;
}

export function BatchImport({ onImport, currentMessages }: BatchImportProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<BatchImportResult | null>(null);
  const [selectedFile, setSelectedFile] = useState<string>("");

  const handleSelectFile = async () => {
    try {
      const filePath = await SelectLogFile();
      if (!filePath) return;

      setSelectedFile(filePath);
      setIsLoading(true);

      const importResult = await ImportLogFile(filePath);
      setResult(importResult);
      setIsLoading(false);
    } catch (error) {
      setIsLoading(false);
      toast({
        title: "✗ Import Error",
        description: error instanceof Error ? error.message : "Failed to import file",
        variant: "destructive",
      });
    }
  };

  const handleImport = (mode: "replace" | "append") => {
    if (!result || result.messages.length === 0) return;

    let finalMessages: string[];
    if (mode === "append" && currentMessages) {
      const existing = currentMessages.trim();
      finalMessages = existing
        ? [...existing.split("\n"), ...result.messages]
        : result.messages;
    } else {
      finalMessages = result.messages;
    }

    onImport(finalMessages);

    toast({
      title: "✓ Messages Imported",
      description: `${result.messages.length} messages ${mode === "append" ? "appended" : "loaded"}`,
      variant: "default",
    });

    // Reset state
    setIsOpen(false);
    setResult(null);
    setSelectedFile("");
  };

  const handleClose = () => {
    setIsOpen(false);
    setResult(null);
    setSelectedFile("");
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="h-9 w-9"
          title="Import from File"
          onClick={() => setIsOpen(true)}
        >
          <FileUpIcon className="w-4 h-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileUpIcon className="w-5 h-5" />
            Batch Import
          </DialogTitle>
          <DialogDescription>
            Import log messages from CSV, JSON, or text files
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* File Selection */}
          <div className="space-y-2">
            <Button
              type="button"
              variant="outline"
              className="w-full justify-start gap-2"
              onClick={handleSelectFile}
              disabled={isLoading}
            >
              {isLoading ? (
                <Loader2Icon className="w-4 h-4 animate-spin" />
              ) : (
                <FileText className="w-4 h-4" />
              )}
              {selectedFile
                ? selectedFile.split("/").pop()
                : "Select log file..."}
            </Button>

            <p className="text-xs text-muted-foreground">
              Supported formats: CSV, JSON, TXT, LOG
            </p>
          </div>

          {/* Results Preview */}
          {result && (
            <div className="space-y-3">
              {/* Stats */}
              <div className="grid grid-cols-3 gap-2 text-center">
                <div className="p-2 bg-muted rounded-lg">
                  <div className="text-lg font-bold">{result.totalLines}</div>
                  <div className="text-[10px] text-muted-foreground">
                    Total Lines
                  </div>
                </div>
                <div className="p-2 bg-green-500/10 border border-green-500/20 rounded-lg">
                  <div className="text-lg font-bold text-green-600">
                    {result.messages.length}
                  </div>
                  <div className="text-[10px] text-muted-foreground">
                    Valid Messages
                  </div>
                </div>
                <div className="p-2 bg-red-500/10 border border-red-500/20 rounded-lg">
                  <div className="text-lg font-bold text-red-600">
                    {result.errors.length}
                  </div>
                  <div className="text-[10px] text-muted-foreground">Errors</div>
                </div>
              </div>

              {/* Messages Preview */}
              {result.messages.length > 0 && (
                <div className="space-y-1">
                  <div className="text-xs font-medium flex items-center gap-1">
                    <CheckCircle2Icon className="w-3 h-3 text-green-500" />
                    Preview (first 5 messages):
                  </div>
                  <div className="bg-muted rounded-lg p-2 max-h-32 overflow-auto">
                    {result.messages.slice(0, 5).map((msg, i) => (
                      <div
                        key={i}
                        className="text-xs font-mono text-muted-foreground truncate py-0.5"
                      >
                        {i + 1}. {msg.length > 80 ? msg.slice(0, 80) + "..." : msg}
                      </div>
                    ))}
                    {result.messages.length > 5 && (
                      <div className="text-xs text-muted-foreground italic pt-1">
                        ... and {result.messages.length - 5} more
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Errors Preview */}
              {result.errors.length > 0 && (
                <div className="space-y-1">
                  <div className="text-xs font-medium flex items-center gap-1 text-red-500">
                    <AlertCircleIcon className="w-3 h-3" />
                    Errors (first 3):
                  </div>
                  <div className="bg-red-500/10 rounded-lg p-2 max-h-20 overflow-auto">
                    {result.errors.slice(0, 3).map((err, i) => (
                      <div key={i} className="text-xs text-red-600 py-0.5">
                        {err}
                      </div>
                    ))}
                    {result.errors.length > 3 && (
                      <div className="text-xs text-red-400 italic pt-1">
                        ... and {result.errors.length - 3} more errors
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          {result && result.messages.length > 0 && (
            <>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleImport("append")}
              >
                Append
              </Button>
              <Button type="button" onClick={() => handleImport("replace")}>
                Replace All
              </Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
