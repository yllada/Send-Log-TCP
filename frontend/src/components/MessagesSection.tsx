import React from "react";

const MessagesSection: React.FC<{ messages: string[]; onChange: (index: number, value: string) => void; onAdd: () => void; onRemove: (index: number) => void; errors: string[] }> = React.memo(({ messages, onChange, onAdd, onRemove, errors }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">Mensajes</label>
    {messages.map((message, index) => (
      <div key={index} className="mb-2 flex items-center">
        <input
          type="text"
          value={message}
          onChange={(e) => onChange(index, e.target.value)}
          className={`flex-grow border ${errors[index] ? "border-red-500" : "border-gray-300"} p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500`}
          required
          aria-invalid={!!errors[index]}
          aria-describedby={errors[index] ? `message-error-${index}` : undefined}
        />
        <button type="button" onClick={onAdd} className="ml-2 bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition" aria-label="Añadir mensaje">+</button>
        {messages.length > 1 && (
          <button type="button" onClick={() => onRemove(index)} className="ml-2 bg-red-500 text-white p-2 rounded hover:bg-red-600 transition" aria-label="Eliminar mensaje">-</button>
        )}
        {errors[index] && <p id={`message-error-${index}`} className="text-red-500 text-sm mt-1">{errors[index]}</p>}
      </div>
    ))}
  </div>
));

export default MessagesSection;