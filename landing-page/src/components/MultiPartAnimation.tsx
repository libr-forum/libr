import { useEffect, useRef } from 'react';
import { Mail, StickyNote } from 'lucide-react';
import { gsap } from 'gsap';
import { MotionPathPlugin } from 'gsap/MotionPathPlugin';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { PcCase,Shield,Database } from 'lucide-react';

import background from '../assets/network_background-01.svg';

gsap.registerPlugin(MotionPathPlugin, ScrollTrigger);

export default function NetworkAnimation() {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const noteRefs = useRef<SVGSVGElement[]>([]);
  const envelopeRefs = useRef<SVGSVGElement[]>([]);
  const shieldRefs = useRef<SVGSVGElement[]>([]);

  useEffect(() => {
    const ctx = gsap.context(() => {
      const tl = gsap.timeline({
        scrollTrigger: {
          trigger: containerRef.current,
          start: 'top top',
          end: '+=3000',
          scrub: true,
          pin: true,
        },
      });

      // Notes to Shields
      noteRefs.current.forEach((note, i) => {
        const shield = shieldRefs.current[i];
        const color = i === 1 ? 'red' : 'green';

        tl.to(note, {
          motionPath: {
            path: `#path-note-${i}`,
            align: `#path-note-${i}`,
            alignOrigin: [0.5, 0.5],
            autoRotate: false,
          },
          duration: 1,
          ease: 'none',
          onUpdate: function () {
            const progress = this.progress();
            if (progress > 0.5) {
              note.style.fill = color;
              if (shield) shield.style.fill = color;
            }
          },
          onComplete: () => {
            if (shield) shield.style.fill = 'black';
          },
        }, 0);

        tl.to(note, {
          motionPath: {
            path: `#path-note-${i}`,
            align: `#path-note-${i}`,
            alignOrigin: [0.5, 0.5],
            autoRotate: false,
            start: 1,
            end: 0,
          },
          duration: 1,
          ease: 'none',
        }, 1);
      });

      // Envelopes to DBs
      envelopeRefs.current.forEach((env, i) => {
        tl.to(env, {
          motionPath: {
            path: `#path-env-${i}`,
            align: `#path-env-${i}`,
            alignOrigin: [0.5, 0.5],
            autoRotate: false,
          },
          duration: 1,
          ease: 'none',
        }, 2);
      });
    }, containerRef);

    return () => ctx.revert();
  }, []);

  return (
    <div
      ref={containerRef}
      className="relative w-full h-[300vh] bg-white"
      style={{
        backgroundImage: `url(${background})`,
        backgroundSize: 'contain',
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center',
      }}
    >
      {/* Nodes */}
      <PcCase x={400} y={300} fill="black" />
      <Shield ref={(el: any) => (shieldRefs.current[0] = el)} x={200} y={100} fill="black" />
      <Shield ref={(el: any) => (shieldRefs.current[1] = el)} x={500} y={120} fill="black" />
      <Shield ref={(el: any) => (shieldRefs.current[2] = el)} x={600} y={450} fill="black" />
      <Database x={100} y={500} fill="black" />
      <Database x={300} y={550} fill="black" />
      <Database x={700} y={500} fill="black" />
      <Database x={800} y={300} fill="black" />

      {/* Moving Notes */}
      {[0, 1, 2].map((i) => (
        <StickyNote
          key={i}
          ref={(el: any) => (noteRefs.current[i] = el)}
          className="absolute w-6 h-6"
          style={{ top: 300, left: 400, fill: 'black' }}
        />
      ))}

      {/* Moving Envelopes */}
      {[0, 1, 2, 3].map((i) => (
        <Mail
          key={i}
          ref={(el: any) => (envelopeRefs.current[i] = el)}
          className="absolute w-6 h-6"
          style={{ top: 300, left: 400, fill: 'black' }}
        />
      ))}

      {/* Curved Quadratic Paths */}
      <svg className="absolute inset-0 w-full h-full pointer-events-none">
        {[0, 1, 2].map((i) => (
          <path
            key={`note-path-${i}`}
            id={`path-note-${i}`}
            d={`M400,300 Q${300 + i * 50},${200 + i * 50} ${200 + i * 150},${100 + i * 100}`}
            stroke="gray"
            fill="none"
            strokeWidth="0.5"
          />
        ))}
        {[0, 1, 2, 3].map((i) => (
          <path
            key={`env-path-${i}`}
            id={`path-env-${i}`}
            d={`M400,300 Q${350 + i * 50},${400 - i * 30} ${100 + i * 200},${500 - i * 100}`}
            stroke="gray"
            fill="none"
            strokeWidth="0.5"
          />
        ))}
      </svg>
    </div>
  );
}
