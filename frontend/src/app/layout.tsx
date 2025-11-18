import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { Toaster } from "@/components/ui/toaster";
import { ThemeProvider } from "@/components/theme-provider";
import { ModeToggle } from "@/components/modeToggle";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "SendLog SysLog",
  description: "Program created to link logs to a syslog server using TCP/UDP",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning className="h-full">
      <body className={`${inter.className} h-full overflow-hidden`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <div className="absolute top-2 right-2 z-50">
            <ModeToggle />
          </div>
          <main className="flex flex-col h-full bg-background overflow-hidden">
            <div className="w-full flex flex-col items-center pt-3 pb-2 px-4 flex-shrink-0">
              <h2 className="text-xl font-bold text-center bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
                SendLog Syslog
              </h2>
              <p className="text-xs text-muted-foreground text-center">
                Connect and send logs to syslog servers
              </p>
            </div>
            <div className="flex-1 w-full overflow-y-auto scrollbar-thin px-4 pb-3 min-h-0">
              {children}
            </div>
          </main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
