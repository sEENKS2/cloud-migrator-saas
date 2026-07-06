import { useState, useEffect } from 'react';

function App() {
  const [progreso, setProgreso] = useState({
    total_registros: 0,
    procesados: 0,
    lotes_fallidos: 0,
    porcentaje: 0,
    status: 'IDLE'
  });

  // Efecto para consultar el progreso continuamente si está procesando
  useEffect(() => {
    let intervalo = null;

    if (progreso.status === 'PROCESSING') {
      intervalo = setInterval(async () => {
        try {
          const res = await fetch('http://localhost:8080/api/progreso');
          const data = await res.json();
          setProgreso(data);

          if (data.status === 'COMPLETED' || data.status === 'FAILED') {
            clearInterval(intervalo);
          }
        } catch (err) {
          console.error("Error consultando el progreso:", err);
        }
      }, 200); // Consulta cada 200 milisegundos por la velocidad de Go
    }

    return () => clearInterval(intervalo);
  }, [progreso.status]);

  const dispararMigracion = async () => {
    try {
      setProgreso((prev) => ({ ...prev, status: 'PROCESSING', porcentaje: 0, procesados: 0 }));
      
      const res = await fetch('http://localhost:8080/api/migrar', { method: 'POST' });
      const data = await res.json();
      console.log(data.mensaje);
    } catch (err) {
      console.error("Error al iniciar la migración:", err);
      setProgreso((prev) => ({ ...prev, status: 'IDLE' }));
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center p-6">
      <div className="max-w-xl w-full bg-gray-800 rounded-xl p-8 shadow-2xl border border-gray-700">
        <h1 className="text-3xl font-bold mb-2 text-center text-indigo-400">Cloud Migrator SaaS</h1>
        <p className="text-gray-400 text-sm text-center mb-8">Migración Masiva de Datos e Integridad en Tiempo Real</p>

        {/* Panel de Estado */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          <div className="bg-gray-700/50 p-4 rounded-lg text-center">
            <span className="text-xs text-gray-400 block uppercase font-semibold">Estado</span>
            <span className={`text-sm font-bold ${progreso.status === 'PROCESSING' ? 'text-yellow-400 animate-pulse' : progreso.status === 'COMPLETED' ? 'text-green-400' : 'text-gray-300'}`}>
              {progreso.status}
            </span>
          </div>
          <div className="bg-gray-700/50 p-4 rounded-lg text-center">
            <span className="text-xs text-gray-400 block uppercase font-semibold">Procesados</span>
            <span className="text-lg font-mono font-bold text-indigo-300">{progreso.procesados.toLocaleString()}</span>
          </div>
          <div className="bg-gray-700/50 p-4 rounded-lg text-center border border-red-500/20">
            <span className="text-xs text-red-400 block uppercase font-semibold">Lotes Fallidos</span>
            <span className="text-lg font-mono font-bold text-red-400">{progreso.lotes_fallidos}</span>
          </div>
        </div>

        {/* Barra de Progreso */}
        <div className="mb-8">
          <div className="flex justify-between text-sm mb-2">
            <span className="text-gray-400">Progreso de la Migración</span>
            <span className="font-bold text-indigo-400">{Math.round(progreso.porcentaje)}%</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-4 overflow-hidden">
            <div 
              className="bg-gradient-to-r from-indigo-500 to-purple-600 h-full transition-all duration-150 ease-out"
              style={{ width: `${progreso.porcentaje}%` }}
            ></div>
          </div>
        </div>

        {/* Acciones */}
        <div className="space-y-4">
          <button
            onClick={dispararMigracion}
            disabled={progreso.status === 'PROCESSING'}
            className={`w-full py-3 px-6 rounded-lg font-bold text-lg transition-all ${
              progreso.status === 'PROCESSING' 
                ? 'bg-gray-600 text-gray-400 cursor-not-allowed' 
                : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/30 active:scale-95'
            }`}
          >
            {progreso.status === 'PROCESSING' ? 'Migrando Datos...' : 'Iniciar Migración Masiva'}
          </button>

          {/* Botón de descarga de errores: solo aparece si terminó y hubo fallas */}
          {progreso.status === 'COMPLETED' && progreso.lotes_fallidos > 0 && (
            <a
              href="http://localhost:8080/api/errores/descargar"
              download
              className="block w-full text-center py-2 px-6 rounded-lg font-semibold text-sm bg-red-950/40 hover:bg-red-900/60 text-red-400 border border-red-500/30 transition-all active:scale-95 shadow-md"
            >
              ⚠️ Descargar Reporte de Errores (.log)
            </a>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;