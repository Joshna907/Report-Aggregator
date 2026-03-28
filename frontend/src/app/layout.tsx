import type { Metadata } from "next";
import "./globals.css";
import { TopNav } from "@/components/TopNav";

export const metadata: Metadata = {
  title: "FOSSology Report Aggregator",
  description: "Aggregating SPDX and CycloneDX into a single true SBOM",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased min-h-screen flex flex-col">
        <TopNav />
        <main className="flex-1 w-full max-w-[1400px] mx-auto p-6">
          {children}
        </main>
      </body>
    </html>
  );
}
