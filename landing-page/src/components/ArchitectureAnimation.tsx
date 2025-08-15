import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import { gsap } from 'gsap';
import { MotionPathPlugin } from 'gsap/MotionPathPlugin';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { ShieldCheck, Monitor, Database } from 'lucide-react';
import {IconMail,IconFileText} from '@tabler/icons-react';

gsap.registerPlugin(MotionPathPlugin, ScrollTrigger);

export default function ArchitectureAnimation() {
  const sectionRef = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(false);

  const [isDarkMode, setIsDarkMode] = useState(false);
  useEffect(() => {
    // Detect theme from <html class="dark">
    const isDark = document.documentElement.classList.contains('dark');
    setIsDarkMode(isDark);

    const observer = new MutationObserver(() => {
      const updatedIsDark = document.documentElement.classList.contains('dark');
      setIsDarkMode(updatedIsDark);
    });

    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

    return () => observer.disconnect();
  }, []);

  const path1Ref = useRef<SVGPathElement>(null);
  const path2Ref = useRef<SVGPathElement>(null);
  const path3Ref = useRef<SVGPathElement>(null);
  const path4Ref = useRef<SVGPathElement>(null);
  const path5Ref = useRef<SVGPathElement>(null);
  const path6Ref = useRef<SVGPathElement>(null);
  const path7Ref = useRef<SVGPathElement>(null);
  const path8Ref = useRef<SVGPathElement>(null);
  const path9Ref = useRef<SVGPathElement>(null);
  const path10Ref = useRef<SVGPathElement>(null);

  const mail1Ref = useRef(null);
  const mail2Ref = useRef(null);
  const mail3Ref = useRef(null);

  const mail4Ref = useRef(null);
  const mail5Ref = useRef(null);
  const mail6Ref = useRef(null);
  const packet1Ref = useRef(null);
  const packet2Ref = useRef(null);
  const packet3Ref = useRef(null);
  const packet4Ref = useRef(null);

  const [scale, setScale] = useState(1);

  const pathRefs = useRef<SVGPathElement[]>([]);
  const pathD = [
    "M761 558 Q838 380 903 548",
    "M761 558 Q723 430 658 465",
    "M761 558 Q813 480 846 640",
    "M903 548 Q838 380 761 558",
    "M658 465 Q723 430 761 558",
    "M846 640 Q813 480 761 558",
    "M761 558 Q813 480 835 502",
    "M761 558 Q830 520 792 687",
    "M761 558 Q700 510 666 627",
    "M761 558 Q770 470 753 455",
  ];

  const circleRefs = useRef<SVGCircleElement[]>([]);
  const circlePositions = [
    { cx: 761, cy: 558 },

    { cx: 846, cy: 640 },
    { cx: 658, cy: 465 },
    { cx: 903, cy: 548 },

    // { cx: 866, cy: 426 },
    { cx: 621, cy: 534 },

    { cx: 835, cy: 502 },
    { cx: 792, cy: 687 },
    { cx: 666, cy: 627 },
    { cx: 753, cy: 455 },
  ];

  useEffect(() => {
    const stroke = isDarkMode ? '#fdfcf7' : '#304a78';
    pathRefs.current.forEach((p) => {
      if (p) p.setAttribute('stroke', stroke);
    });
  }, [isDarkMode]);

  useEffect(() => {
    const fill = isDarkMode ? '#fdfcf7' : '#304a78';
    circleRefs.current.forEach(circle => {
      if (circle) circle.setAttribute('fill', fill);
    });
  }, [isDarkMode]);

  
  useEffect(() => {
    const ctx = gsap.context(() => {
      if (
        !path1Ref.current || !path2Ref.current || !path3Ref.current ||
        !path4Ref.current || !path5Ref.current || !path6Ref.current || !path7Ref.current || !path8Ref.current || !path9Ref.current || !path10Ref.current
      ) return;

      gsap.set(
        [
          mail1Ref.current, mail2Ref.current, mail3Ref.current,mail4Ref.current,mail5Ref.current,mail6Ref.current,
          packet1Ref.current, packet2Ref.current, packet3Ref.current, packet4Ref.current
        ],
        { xPercent: -50, yPercent: -50 }
      );

      const masterTimeline = gsap.timeline({ repeat: -1 });

      // Mail timeline
      const mailTimeline = gsap.timeline({
        defaults: { duration: 1.5, ease: 'linear' },
      });

      mailTimeline.to(mail1Ref.current, {
        motionPath: {
          path: path1Ref.current,
          align: path1Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(mail2Ref.current, {
        motionPath: {
          path: path2Ref.current,
          align: path2Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(mail3Ref.current, {
        motionPath: {
          path: path3Ref.current,
          align: path3Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0);

      // Return timeline
      const returnTimeline = gsap.timeline({
        defaults: { duration: 2, ease: 'linear' },
      });

      returnTimeline.to(mail4Ref.current, {
        motionPath: {
          path: path4Ref.current,
          align: path4Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(mail5Ref.current, {
        motionPath: {
          path: path5Ref.current,
          align: path5Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0.4).to(mail6Ref.current, {
        motionPath: {
          path: path6Ref.current,
          align: path6Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0.5);

      // Packet timeline
      const packetTimeline = gsap.timeline({
        defaults: { duration: 1.5, ease: 'linear' }
      });

      packetTimeline.to(packet1Ref.current, {
        motionPath: {
          path: path7Ref.current,
          align: path7Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(packet2Ref.current, {
        motionPath: {
          path: path8Ref.current,
          align: path8Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(packet3Ref.current, {
        motionPath: {
          path: path9Ref.current,
          align: path9Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0).to(packet4Ref.current, {
        motionPath: {
          path: path10Ref.current,
          align: path10Ref.current,
          alignOrigin: [0.5, 0.5],
        }
      }, 0);

      const resetPositions = () => {
        gsap.set([
          mail1Ref.current, mail2Ref.current, mail3Ref.current,mail4Ref.current,mail5Ref.current,mail6Ref.current,
          packet1Ref.current, packet2Ref.current, packet3Ref.current, packet4Ref.current
        ], { clearProps: 'all', xPercent: -50, yPercent: -50 });
      };

      masterTimeline
        .add(() => resetPositions(), 0)
        .add(mailTimeline,"+=2")
        .add(returnTimeline,"+=0.5")
        .add(packetTimeline,"+=0.5");

    }, sectionRef);

    return () => ctx.revert();
  }, []);

  useEffect(() => {
    function handleResize() {
      // Base SVG size
      const baseWidth = 1536;
      const baseHeight = 1024;
      const minWidth = 900;

      // Get current window or section size
      const w = Math.max(window.innerWidth, minWidth);
      const h = window.innerHeight;

      // Decide factor based on width breakpoints
      let factor = 1.5;
      // if (w < 1750 && w >= 1400) {
      //   factor = 1.5;
      // } else if (w < 1400) {
      //   factor = 1.5;
      // }

      // Calculate scale to fit both width and height, keeping aspect ratio
      const scaleW = w / baseWidth;
      const scaleH = h / baseHeight;

      // Use the smaller scale to ensure it fits
      setScale(Math.min(scaleW * factor, scaleH * factor, factor));
    }
    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => setIsVisible(entry.isIntersecting),
      { threshold: 0.2 }
    );
    if (sectionRef.current) observer.observe(sectionRef.current);
    return () => observer.disconnect();
  }, []);
  
  const [width, setWidth] = useState(window.innerWidth);
  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);
  const isSmall=width<1200;

  const [yPos, setYPos] = useState(0);
  useEffect(() => {
    function updateY() {
      const w = window.innerWidth;
      const isDesktop = w >= 1200;

      if (!isDesktop) {
        // --- MOBILE/TABLET LOGIC ---
        if (w < 400) {
          setYPos(-200);
        } else if (w < 768) {
          setYPos(-300);
        } else if (w < 1400) {
          // Linear interpolation from -300 at 768px to -250 at 1400px
          const ratio = (w - 768) / (1400 - 768);
          const yValue = -300 + ratio * (50); // -300 to -250
          setYPos(yValue);
        } else {
          setYPos(1000);
        }
      } else {
        // --- DESKTOP LOGIC PLACEHOLDER ---
        // TODO: Implement desktop-specific shifting logic here
        // Example: setYPos(-100);
        setYPos(-100); // Placeholder value for desktop
      }
    }
    updateY(); // Initial run
    window.addEventListener("resize", updateY);
    return () => window.removeEventListener("resize", updateY);
  }, []);

  return (
    <section
      ref={sectionRef}
      className="relative h-[80vh] w-full flex items-center justify-center [&_*]:box-border [&_*]:break-words [&_*]:whitespace-normal"
    >
      {/* Background */}
      {/* <div
        className="absolute inset-0 w-full h-full pointer-events-none flex items-center justify-center opacity-50 z-0 p-28"
        style={{backgroundImage: `url(${isDarkMode ? WorkingBGDark : WorkingBGLight})`,}}
      /> */}
      <motion.div
        className="relative origin-center"
        style={{
          transformOrigin: 'center',
        }}
        initial={{ opacity: 0, y: 0, scale: scale }}
        animate={{
          opacity: isVisible ? 1 : 0,
          y: isVisible ? yPos : 40,
          scale,
        }}
        transition={{
          duration: 0.8,
          ease: [0.4, 0, 0.2, 1],
        }}
      >
        {/* Foreground Animation */}
        <svg
          width="1536"
          height="1024"
          viewBox="0 60 1536 1024"
          className="rounded z-10"
        >
          {/* Paths */}
          <path ref={path1Ref} d="M761 558 Q838 380 903 548" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path2Ref} d="M761 558 Q723 430 658 465" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path3Ref} d="M761 558 Q813 480 846 640" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path4Ref} d="M903 548 Q838 380 761 558" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path5Ref} d="M658 465 Q723 430 761 558" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path6Ref} d="M846 640 Q813 480 761 558" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path7Ref} d="M761 558 Q813 480 835 502" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path8Ref} d="M761 558 Q830 520 792 687" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path9Ref} d="M761 558 Q700 510 666 627" stroke="black" strokeWidth="0.5" fill="none" />
          <path ref={path10Ref} d="M761 558 Q770 470 753 455" stroke="black" strokeWidth="0.5" fill="none" />
          {pathD.map((d, i) => (
            <path
              key={i}
              ref={(el) => (pathRefs.current[i] = el!)}
              d={d}
              strokeWidth="0.5"
              fill="none"
              style={{ transition: 'stroke 0.3s ease-in-out' }}
            />
          ))}
          {/* Mails */}
          <IconFileText ref={mail1Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconFileText ref={mail2Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconFileText ref={mail3Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconFileText ref={mail4Ref} className="text-green-500 fill-libr-primary"/>
          <IconFileText ref={mail5Ref} className="text-green-500 fill-libr-primary"/>
          <IconFileText ref={mail6Ref} className="text-red-500 fill-libr-primary"/>
          {/* Packets */}
          <IconMail ref={packet1Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconMail ref={packet2Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconMail ref={packet3Ref} className="text-libr-secondary fill-libr-primary"/>
          <IconMail ref={packet4Ref} className="text-libr-secondary fill-libr-primary"/>
          {/* Optional Static Circles */}
          {circlePositions.map((pos, i) => (
            <circle
              key={i}
              ref={(el) => (circleRefs.current[i] = el!)}
              cx={pos.cx}
              cy={pos.cy}
              r="20"
              opacity={1}
              style={{ transition: 'fill 0.3s ease-in-out' }}
            />
          ))}
        </svg>

        <Monitor style={{ left: 761, top: 498 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />

        <ShieldCheck style={{ left: 846, top: 580 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />
        <ShieldCheck style={{ left: 658, top: 405 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />
        <ShieldCheck style={{ left: 903, top: 488 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />

        {/* <Monitor style={{ left: 866, top: 366 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" /> */}
        <Monitor style={{ left: 621, top: 474 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />

        <Database style={{ left: 835, top: 442 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />
        <Database style={{ left: 792, top: 627 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />
        <Database style={{ left: 666, top: 567 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />
        <Database style={{ left: 753, top: 395 }} className="absolute w-6 h-6 text-libr-primary -translate-x-1/2 -translate-y-1/2 z-20" />

        <motion.div
          className="absolute libr-card p-4 text-left min-w-[230px]"
          style={{
            left: isSmall ? 660 : 960,
            top: isSmall ? 700 : 320,
          }}
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.6}}
          viewport={{ once: true }}
          
        >
          <div className='flex flex-row items-center space-x-2 mb-2'>
            <Monitor className="w-6 h-6 text-libr-secondary" />
            <h3 className="text-lg font-semibold text-libr-secondary">Client</h3>
          </div>
          <p className="text-muted-foreground text-xs">Send a message<br/>Wait for moderators<br/>Aggregate responses, sign and<br/>send to database if approved</p>
        </motion.div>
        <motion.div
          className="absolute libr-card p-4 text-left"
          style={{
            left: isSmall ? 660 : 360,
            top: isSmall ? 990 : 420,
            minWidth: width < 1400 ? 230 : undefined,
          }}
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.6}}
          viewport={{ once: true }}
        >
          <div className='flex flex-row items-center space-x-2 mb-2'>
            <Database className="w-6 h-6 text-libr-secondary" />
            <h3 className="text-lg font-semibold text-libr-secondary">Database</h3>
          </div>
          <p className="text-muted-foreground text-xs">Uses timestamp hash to<br/>calculate storage nodes<br/>Verify signatures<br/>Store with replication</p>
        </motion.div>
        <motion.div
          className="absolute libr-card p-4 text-left min-w-[230px]"
          style={{
            left: isSmall ? 660 : 960,
            top: isSmall ? 850 : 570,
          }}
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.6}}
          viewport={{ once: true }}
        >
          <div className='flex flex-row items-center space-x-2 mb-2'>
            <ShieldCheck className="w-6 h-6 text-libr-secondary" />
            <h3 className="text-lg font-semibold text-libr-secondary">Moderator</h3>
          </div>
          <p className="text-muted-foreground text-xs">Recieve new messages<br/>Moderate per community rules<br/>Sign and return to client</p>
        </motion.div>
      </motion.div>
    </section>
  );
}
