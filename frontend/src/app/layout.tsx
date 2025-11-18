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
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <div className="absolute top-2 right-2 z-50">
            <ModeToggle />
          </div>
          <main className="flex flex-col items-center min-h-screen p-6">
            <h2 className="text-2xl font-bold text-center mb-3 bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
              SendLog Syslog
            </h2>
            <p className="text-sm text-muted-foreground text-center mb-4">
              Connect and send logs to syslog servers
            </p>
            {children}
          </main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
