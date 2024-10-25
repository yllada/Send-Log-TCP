import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { Toaster } from "@/components/ui/toaster";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { ModeToggle } from "@/components/modeToggle";

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
          <main className="flex flex-col items-center min-h-screen bg-gradient-to-b from-gray-100 to-gray-300 p-6">
            <h2
              className="text-3xl font-bold text-center text-gray-800 mb-4 
               bg-gradient-to-r from-blue-500 to-purple-500 
               bg-clip-text text-transparent"
            >
              SendLog Syslog
            </h2>
            <p className="text-lg text-gray-700 text-center mt-2 mb-6">
              A tool to link logs to a syslog server using TCP/UDP
            </p>
            {children}
          </main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
