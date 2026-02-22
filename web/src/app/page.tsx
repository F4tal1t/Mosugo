"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import gsap from "gsap";

function WindowsIcon() {
  return (
    <Image src="/window.svg" alt="Windows icon" width={24} height={24} />
  );
}

function DownloadIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
      <path d="M21 15v4h-2v-4zm-2 4v2H5v-2zM5 15v4H3v-4zm8-12v14h-2V3z" />
      <path d="M7 11v2h10v-2zm2 2v2h2v-2zm4 0v2h2v-2z" />
      <path d="M15 11v2h2v-2z" />
    </svg>
  );
}

function GitBranchIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
      <path d="M4 14h4v2H4zm0 6h4v2H4zm-2-4h2v4H2zm6 0h2v4H8zm8-14h4v2h-4zm0 6h4v2h-4zm-2-4h2v4h-2zm6 0h2v4h-2zm-8 13h5v2h-5zm5-5h2v5h-2zM5 2h2v10H5z" />
    </svg>
  );
}

export default function Home() {
  const containerRef = useRef<HTMLDivElement>(null);
  const appWindowRef = useRef<HTMLDivElement>(null);
  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    // Fetch GitHub stars
    fetch("https://api.github.com/repos/F4tal1t/Mosugo")
      .then((res) => res.json())
      .then((data) => {
        if (data.stargazers_count !== undefined) {
          setStars(data.stargazers_count);
        }
      })
      .catch(() => { });

    const ctx = gsap.context(() => {
      // Elements targeting
      const t1 = gsap.timeline({ defaults: { ease: "power4.out" } });

      t1.fromTo(
        ".nav-item",
        { y: -30, opacity: 0 },
        { y: 0, opacity: 1, duration: 1, stagger: 0.1 }
      )
        .fromTo(
          ".hero-text",
          { y: 30, opacity: 0 },
          { y: 0, opacity: 1, duration: 1.2, stagger: 0.15 },
          "-=0.6"
        )
        .fromTo(
          ".hero-button",
          { y: 20, opacity: 0, scale: 0.95 },
          { y: 0, opacity: 1, scale: 1, duration: 1, ease: "back.out(1.7)" },
          "-=0.8"
        );

      // Ensure 3D context is ready
      gsap.set(appWindowRef.current, { perspective: 1000 });
      gsap.set(".app-window-inner", { transformStyle: "preserve-3d" });

    }, containerRef);

    return () => ctx.revert();
  }, []);

  // Magnetic button effect handler
  const handleMagneticMove = (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => {
    const element = e.currentTarget;
    const rect = element.getBoundingClientRect();
    const x = e.clientX - rect.left - rect.width / 2;
    const y = e.clientY - rect.top - rect.height / 2;

    gsap.to(element, {
      x: x * 0.15,
      y: y * 0.15,
      duration: 0.4,
      ease: "power2.out"
    });
  };

  const handleMagneticLeave = (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => {
    const element = e.currentTarget;
    gsap.to(element, {
      x: 0,
      y: 0,
      duration: 0.7,
      ease: "elastic.out(1, 0.3)"
    });
  };

  // 3D Tilt effect handler for the App Window
  const handleWindowMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!appWindowRef.current) return;
    const rect = appWindowRef.current.getBoundingClientRect();

    // Calculate relative mouse position (0 to 1) based on the center of the element
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const xMid = rect.width / 2;
    const yMid = rect.height / 2;

    // Calculate rotation (-1 to 1 multiplied by max rotation degrees)
    const rotationX = ((y - yMid) / yMid) * -10; // Max tilt up/down (10 deg)
    const rotationY = ((x - xMid) / xMid) * 10;  // Max tilt left/right (10 deg)

    gsap.to(".app-window-inner", {
      rotationX: rotationX,
      rotationY: rotationY,
      duration: 0.5,
      ease: "power2.out",
    });
  };

  const handleWindowLeave = () => {
    gsap.to(".app-window-inner", {
      rotationX: 0,
      rotationY: 0,
      duration: 0.8,
      ease: "elastic.out(1, 0.5)",
    });
  };

  return (
    <main ref={containerRef} className="relative h-screen w-full overflow-hidden flex flex-col font-sans">

      {/* Navbar */}
      <nav className="w-full max-w-7xl mx-auto px-8 py-8 flex justify-between items-center z-50">
        <div className="nav-item font-bold text-3xl tracking-tighter mix-blend-difference flex items-center gap-3">
          Mosugo
          <Image
            src="/Mosugo_Icon.png"
            alt="Mosugo Logo"
            width={40}
            height={40}
            className="rounded-xl border border-[var(--color-mosugo-accent)]/20 p-1.5 bg-mosugo-bg-light shadow-sm"
          />
        </div>
        <a
          href="https://github.com/F4tal1t/Mosugo"
          target="_blank"
          rel="noopener noreferrer"
          onMouseMove={handleMagneticMove}
          onMouseLeave={handleMagneticLeave}
          className="nav-item group relative flex items-center gap-2 px-6 py-3 rounded-full overflow-hidden border border-[var(--color-mosugo-accent)]/20 hover:border-[var(--color-mosugo-accent)] transition-colors"
        >
          <div className="absolute inset-0 bg-[var(--color-mosugo-accent)] translate-y-[100%] group-hover:translate-y-0 transition-transform duration-500 ease-[cubic-bezier(0.19,1,0.22,1)]" />
          <span className="relative z-10 text-[var(--color-mosugo-accent)] group-hover:text-white transition-colors duration-300 font-medium text-sm flex items-center gap-2">
            <GitBranchIcon /> Star on GitHub {stars !== null && <span className="opacity-60 text-xs ml-1 font-bold">{stars}</span>}
          </span>
        </a>
      </nav>

      {/* Hero Section - Two Column Layout */}
      <section className="flex-1 w-full max-w-[90rem] mx-auto px-8 flex flex-col lg:flex-row items-center justify-between z-10 gap-16 relative pb-12">

        {/* Left Column: Text Content */}
        <div className="w-full lg:w-[45%] flex flex-col items-start pt-10 lg:pt-0">
          <h1 className="hero-text text-[clamp(1.5rem,5vw,4rem)] leading-[0.9] tracking-[-0.03em] font-medium text-balance mb-8">
            Think Spatial
          </h1>
          <h1 className="hero-text text-[var(--color-mosugo-accent)]/50 text-[clamp(1.5rem,5vw,4rem)] leading-[0.9] tracking-[-0.03em] font-medium text-balance mb-8">
            Work Unbound
          </h1>

          <p className="hero-text text-base md:text-lg text-[var(--color-mosugo-accent)]/70 max-w-sm mb-12 leading-relaxed">
            A minimal, infinite canvas application. Combine the freedom of spatial notes with the structure of daily workspaces.
          </p>

          <a
            href="/Mosugo-Setup-1.0.1.exe"
            download
            onMouseMove={handleMagneticMove}
            onMouseLeave={handleMagneticLeave}
            className="hero-button group inline-flex relative overflow-hidden bg-[var(--color-mosugo-accent)] text-white px-8 py-4 rounded-full font-medium text-lg shadow-[0_8px_30px_rgb(10,25,47,0.2)] hover:shadow-[0_8px_30px_rgb(10,25,47,0.4)] transition-shadow"
          >
            <div className="absolute inset-0 bg-white/20 translate-y-[100%] group-hover:translate-y-0 transition-transform duration-500 ease-[cubic-bezier(0.19,1,0.22,1)]" />
            <span className="relative flex items-center gap-3">
              <WindowsIcon />
              Download for Windows
            </span>
          </a>
        </div>

        {/* Right Column: High Fidelity Placeholder with 3D Interaction */}
        <div
          ref={appWindowRef}
          onMouseMove={handleWindowMove}
          onMouseLeave={handleWindowLeave}
          className="w-full lg:w-[55%] h-[50vh] lg:h-[75vh] relative flex items-center justify-center -mr-12 lg:-mr-24"
        >
          <div className="app-window-inner w-full h-full bg-mosugo-accent border border-[var(--color-mosugo-accent)]/15 backdrop-blur-2xl rounded-l-3xl shadow-[-20px_0_60px_rgba(10,25,47,0.08)] flex items-center justify-center overflow-hidden z-20 relative">

            <div className="absolute top-0 left-0 w-full h-10 border-b border-[var(--color-mosugo-accent)]/10 flex items-center justify-between px-4 bg-mosugo-accent z-30">
              <div className="text-xs font-bold text-white/70 flex items-center gap-2">
                <Image src="/Mosugo_Icon.png" alt="Icon" width={20} height={20} />
                Mosugo
              </div>
              <div className="flex gap-5 opacity-40 items-center">
                {/* Minimize */}
                <div className="w-3 h-[1px] bg-white" />
                {/* Maximize */}
                <div className="w-3 h-3 border border-white" />
                {/* Close */}
                <div className="relative w-3 h-3 flex items-center justify-center">
                  <div className="absolute w-3 h-[1px] bg-mosugo-accent rotate-45" />
                  <div className="absolute w-3 h-[1px] bg-mosugo-accent rotate-45" />
                </div>
              </div>
            </div>

            <div className="relative w-full h-full pt-10">
              <Image src="/Img.png" alt="App UI" fill className="object-cover object-left-top" />
            </div>

            {/* Ambient Glow behind the window inside */}
            <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-3/4 h-3/4 bg-[var(--color-mosugo-accent)]/5 blur-[100px] rounded-full pointer-events-none z-0" />
          </div>
        </div>

      </section>

    </main>
  );
}
