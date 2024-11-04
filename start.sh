#!/bin/bash

# Nombre del archivo ejecutable de tu programa Go
EXECUTABLE="mistral-tg"

# Ruta al archivo donde se guardará el ID de proceso (PID)
PID_FILE="$EXECUTABLE.pid"

# Comprobar si el programa ya se está ejecutando
if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
    echo "El programa ya se está ejecutando (PID $(cat "$PID_FILE"))."
    exit 1
fi

# Compilar el programa Go si el ejecutable no existe
if [ ! -f "$EXECUTABLE" ]; then
    echo "Compilando el programa Go..."
    go build -o "$EXECUTABLE" || { echo "Error al compilar el programa."; exit 1; }
fi

# Ejecutar el programa en segundo plano usando nohup y guardar el PID
echo "Iniciando el programa en segundo plano..."
nohup ./"$EXECUTABLE" > "$EXECUTABLE.log" 2>&1 &

# Guardar el PID en el archivo .pid
echo $! > "$PID_FILE"
echo "Programa iniciado con PID $(cat "$PID_FILE")."
