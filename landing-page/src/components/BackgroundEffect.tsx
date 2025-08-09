import React, { useEffect, useRef, useState } from 'react';

export const BackgroundEffect: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isDarkMode, setIsDarkMode] = useState<boolean>(false);
  const [nodes, setNodes] = useState<{ x: number; y: number; radius: number }[]>([]);

  useEffect(() => {
    const isDark = document.documentElement.classList.contains('dark');
    setIsDarkMode(isDark);
    // Generate nodes only once on mount
    if (nodes.length === 0) {
      const width = window.innerWidth;
      const height = window.innerHeight;
      const starsCount = Math.floor(width * height / 12000);
      const generated: { x: number; y: number; radius: number }[] = [];
      for (let i = 0; i < starsCount; i++) {
        const centerX = width / 2;
        const centerY = height / 2;
        let x, y, distanceFromCenter;
        do {
          x = Math.random() * width;
          y = Math.random() * height;
          const dx = (x - centerX) / (width / 2);
          const dy = (y - centerY) / (height / 2);
          distanceFromCenter = Math.sqrt(dx * dx + dy * dy);
        } while (Math.random() > distanceFromCenter * 0.8);
        generated.push({
          x,
          y,
          radius: Math.random() * 1.5 + 3
        });
      }
      setNodes(generated);
    }
    const observer = new MutationObserver(() => {
      const updatedIsDark = document.documentElement.classList.contains('dark');
      setIsDarkMode(updatedIsDark);
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    function drawNetwork() {
      canvas!.width = window.innerWidth;
      canvas!.height = window.innerHeight;
      const starColors = isDarkMode ? [
        'rgba(61, 76, 86, 0.7)',
        'rgba(61, 76, 86, 0.4)',
        'rgba(31, 164, 169, 0.7)',
        'rgba(31, 164, 169, 0.4)'
      ] : [
        'rgba(48, 74, 120, 0.7)',
        'rgba(48, 74, 120, 0.4)',
        'rgba(100, 116, 139, 0.7)',
        'rgba(100, 116, 139, 0.4)'
      ];
      ctx!.clearRect(0, 0, canvas!.width, canvas!.height);
      // Draw nodes
      nodes.forEach((node) => {
        // Pick color based on theme
        const color = starColors[Math.floor(Math.random() * starColors.length)];
        ctx!.save();
        ctx!.beginPath();
        ctx!.arc(node.x, node.y, node.radius, 0, Math.PI * 2);
        ctx!.fillStyle = color;
        ctx!.shadowColor = color;
        ctx!.shadowBlur = 4;
        ctx!.fill();
        ctx!.shadowBlur = 0;
        ctx!.lineWidth = 1;
        ctx!.strokeStyle = 'rgba(255,255,255,0.08)';
        ctx!.stroke();
        ctx!.restore();
      });
      // Draw connections (network style, avoid duplicates)
      const connectionDistance = Math.min(160, canvas!.width * 0.16);
      const drawnPairs = new Set<string>();
      for (let i = 0; i < nodes.length; i++) {
        const distances = [];
        for (let j = 0; j < nodes.length; j++) {
          if (i === j) continue;
          const dx = nodes[i].x - nodes[j].x;
          const dy = nodes[i].y - nodes[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < connectionDistance) {
            distances.push({ idx: j, distance: dist });
          }
        }
        distances.sort(function(a, b) { return a.distance - b.distance; });
        const neighbors = distances.slice(0, 8);
        neighbors.forEach(function(neighbor) {
          const idx = neighbor.idx;
          // Only draw if not already drawn
          const key = i < idx ? `${i},${idx}` : `${idx},${i}`;
          if (!drawnPairs.has(key)) {
            drawnPairs.add(key);
            ctx!.beginPath();
            ctx!.moveTo(nodes[i].x, nodes[i].y);
            ctx!.lineTo(nodes[idx].x, nodes[idx].y);
            ctx!.strokeStyle = 'rgba(180, 200, 220, 0.28)';
            ctx!.lineWidth = 0.7;
            ctx!.stroke();
          }
        });
      }
    }

    drawNetwork();
    window.addEventListener('resize', drawNetwork);
    return () => {
      window.removeEventListener('resize', drawNetwork);
    };
  }, [isDarkMode, nodes]);

  return (
    <canvas
      ref={canvasRef}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        zIndex: -1,
        pointerEvents: 'none'
      }}
    />
  );
}