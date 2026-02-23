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
  openGraph: {
    title: "Mosugo - Spatial Notes Workspace",
    description: "A minimal, spatial notes application built for Windows.",
    url: "https://mosugo.dibby.me",
    siteName: "Mosugo",
    images: [
      {
        url: "/Og.webp",
        width: 1200,
        height: 630,
        alt: "Mosugo Preview Image",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Mosugo - Spatial Notes Workspace",
    description: "A minimal, spatial notes application built for Windows.",
    images: ["/Og.webp"],
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
