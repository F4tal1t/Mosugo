import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";

const comicFont = localFont({
  src: "../../public/fonts/Comic.ttf",
  variable: "--font-comic",
  display: 'swap',
});

export const metadata: Metadata = {
  title: "Mosugo - Spatial Notes Workspace",
  description: "A minimal, spatial notes application built for Windows.",
  icons: {
    icon: "/Mosugo_Icon.ico",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${comicFont.variable}`}>
      <body className="font-sans antialiased">{children}</body>
    </html>
  );
}
