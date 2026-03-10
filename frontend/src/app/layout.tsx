import type { Metadata } from "next";
import { Toaster } from "@/components/ui/toaster";
import { ThemeProvider } from "@/components/theme-provider";
import { ModeToggle } from "@/components/modeToggle";
import { AppMenu } from "@/components/appMenu";
import { Footer } from "@/components/footer";
import "./globals.css";

export const metadata: Metadata = {
  title: "SendLog Syslog",
  description: "Professional syslog testing tool for TCP/UDP connections",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          {/* Mica background layer */}
          <div className="min-h-screen flex flex-col mica-bg">
            {/* Header with Fluent styling */}
            <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/60 backdrop-blur-xl">
              <div className="max-w-5xl mx-auto px-4 h-12 flex justify-between items-center">
                <div className="flex items-center gap-3">
                  <AppMenu />
                </div>
                <ModeToggle />
              </div>
            </header>

            {/* Main content */}
            <main className="flex-1 flex flex-col items-center px-4 py-6">
              {children}
            </main>

            {/* Footer */}
            <Footer />
          </div>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
