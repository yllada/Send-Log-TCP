"use client";

import { useState } from "react";
import {
  Menubar,
  MenubarContent,
  MenubarItem,
  MenubarMenu,
  MenubarSeparator,
  MenubarTrigger,
} from "@/components/ui/menubar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ExternalLink, Info, FileText, LogOut, Github, Code2, Layers } from "lucide-react";
import { Quit } from "@/wailsjs/runtime/runtime";

export function AppMenu() {
  const [showAbout, setShowAbout] = useState(false);

  const handleExit = () => {
    Quit();
  };

  const handleDocumentation = () => {
    window.open("https://github.com/yllada/Send-Log-TCP", "_blank");
  };

  const handleRFCReference = () => {
    window.open("https://datatracker.ietf.org/doc/html/rfc5424", "_blank");
  };

  return (
    <>
      <Menubar className="border-none bg-transparent">
        <MenubarMenu>
          <MenubarTrigger className="text-sm font-normal cursor-pointer">
            File
          </MenubarTrigger>
          <MenubarContent>
            <MenubarItem onClick={handleExit} className="text-sm">
              <LogOut className="w-3.5 h-3.5 mr-2" />
              Exit
            </MenubarItem>
          </MenubarContent>
        </MenubarMenu>

        <MenubarMenu>
          <MenubarTrigger className="text-sm font-normal cursor-pointer">
            Help
          </MenubarTrigger>
          <MenubarContent>
            <MenubarItem onClick={handleDocumentation} className="text-sm">
              <FileText className="w-3.5 h-3.5 mr-2" />
              Documentation
              <ExternalLink className="w-3 h-3 ml-auto opacity-50" />
            </MenubarItem>
            <MenubarItem onClick={handleRFCReference} className="text-sm">
              <FileText className="w-3.5 h-3.5 mr-2" />
              RFC 5424 Reference
              <ExternalLink className="w-3 h-3 ml-auto opacity-50" />
            </MenubarItem>
            <MenubarSeparator />
            <MenubarItem onClick={() => setShowAbout(true)} className="text-sm">
              <Info className="w-3.5 h-3.5 mr-2" />
              About
            </MenubarItem>
          </MenubarContent>
        </MenubarMenu>
      </Menubar>

      <Dialog open={showAbout} onOpenChange={setShowAbout}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold flex items-center gap-2">
              <Code2 className="w-6 h-6" />
              SendLog SysLog
            </DialogTitle>
            <DialogDescription>
              Professional syslog message sender for network testing and monitoring
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Version & License */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-muted-foreground mb-1">Version</p>
                <p className="font-semibold">1.3.0</p>
              </div>
              <div>
                <p className="text-muted-foreground mb-1">License</p>
                <p className="font-semibold">Apache-2.0</p>
              </div>
            </div>

            {/* Technologies */}
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground flex items-center gap-2">
                <Layers className="w-4 h-4" />
                Built With
              </p>
              <div className="grid grid-cols-3 gap-2 text-xs">
                <div className="flex flex-col items-center p-3 border rounded-lg bg-muted/50">
                  <span className="font-semibold">Wails v2</span>
                  <span className="text-muted-foreground">Desktop</span>
                </div>
                <div className="flex flex-col items-center p-3 border rounded-lg bg-muted/50">
                  <span className="font-semibold">Go 1.22</span>
                  <span className="text-muted-foreground">Backend</span>
                </div>
                <div className="flex flex-col items-center p-3 border rounded-lg bg-muted/50">
                  <span className="font-semibold">Next.js 14</span>
                  <span className="text-muted-foreground">Frontend</span>
                </div>
              </div>
            </div>

            {/* Features */}
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Features</p>
              <ul className="text-sm space-y-1 list-disc list-inside">
                <li>RFC 5424 & RFC 3164 syslog formats</li>
                <li>TCP/UDP protocol support with TLS encryption</li>
                <li>RFC 6587 framing methods (Octet Counting & Non-Transparent)</li>
                <li>Real-time connection testing</li>
              </ul>
            </div>

            {/* Author */}
            <div className="pt-2 border-t">
              <div className="flex items-center justify-between text-sm">
                <div>
                  <p className="text-muted-foreground">Authors</p>
                  <p className="font-semibold">Yadian Llada Lopez</p>
                  <p className="font-semibold">José L. Quiñones Rojas</p>
                </div>
                <a
                  href="https://github.com/yllada/Send-Log-TCP"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
                >
                  <Github className="w-5 h-5" />
                  <span className="text-xs">View on GitHub</span>
                  <ExternalLink className="w-3 h-3" />
                </a>
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
