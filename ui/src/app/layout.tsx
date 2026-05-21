import type { Metadata } from "next"
import { Inter } from "next/font/google"
import "./globals.css"
import { Providers } from "./providers"
import { ThemeToggle } from "@/components/shared/theme-toggle"
import { Lightbulb } from "lucide-react"
import Link from "next/link"

const inter = Inter({ subsets: ["latin"], variable: "--font-alliance" })

export const metadata: Metadata = {
  title: "Datum Insights",
  description: "Kubernetes resource insights and policy management",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <Providers>
          <div className="min-h-screen bg-background">
            {/* Header */}
            <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
              <div className="container flex h-14 items-center">
                <div className="flex items-center gap-2 mr-4">
                  <div className="flex items-center justify-center w-8 h-8 rounded-md bg-midnight-fjord text-aurora-moss">
                    <Lightbulb className="h-5 w-5" />
                  </div>
                  <span className="font-semibold text-lg">Datum Insights</span>
                </div>
                <nav className="flex items-center gap-6 text-sm">
                  <Link
                    href="/"
                    className="transition-colors hover:text-foreground/80 text-foreground"
                  >
                    Dashboard
                  </Link>
                  <Link
                    href="/insights"
                    className="transition-colors hover:text-foreground/80 text-muted-foreground"
                  >
                    Insights
                  </Link>
                  <Link
                    href="/policies"
                    className="transition-colors hover:text-foreground/80 text-muted-foreground"
                  >
                    Policies
                  </Link>
                </nav>
                <div className="flex-1" />
                <ThemeToggle />
              </div>
            </header>

            {/* Main Content */}
            <main className="container py-6">{children}</main>

            {/* Footer */}
            <footer className="border-t py-6 md:py-0">
              <div className="container flex flex-col items-center justify-between gap-4 md:h-14 md:flex-row">
                <p className="text-center text-sm text-muted-foreground md:text-left">
                  Built with the Datum Insights API
                </p>
                <p className="text-center text-sm text-muted-foreground md:text-right">
                  <a
                    href="https://datum.net"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline underline-offset-4 hover:text-foreground"
                  >
                    datum.net
                  </a>
                </p>
              </div>
            </footer>
          </div>
        </Providers>
      </body>
    </html>
  )
}
