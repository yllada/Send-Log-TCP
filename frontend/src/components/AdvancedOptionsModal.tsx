import React from "react";
import InputField from "./InputField";

const AdvancedOptionsModal: React.FC<{ facility: number; setFacility: (value: number) => void; severity: number; setSeverity: (value: number) => void; interval: number; setInterval: (value: number) => void; onClose: () => void }> = ({ facility, setFacility, severity, setSeverity, interval, setInterval, onClose }) => (
  <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50">
    <div className="bg-white p-6 rounded-lg shadow-lg w-96">
      <h3 className="text-xl font-bold mb-4">Opciones Avanzadas</h3>
      <InputField label="Facility" value={facility} onChange={(value) => setFacility(Number(value))} type="number" />
      <InputField label="Severity" value={severity} onChange={(value) => setSeverity(Number(value))} type="number" />
      <InputField label="Interval (segundos)" value={interval} onChange={(value) => setInterval(Number(value))} type="number" />
      <button onClick={onClose} className="mt-4 bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition">Cerrar</button>
    </div>
  </div>
);

export default AdvancedOptionsModal;