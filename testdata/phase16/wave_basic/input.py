import io
import struct
import wave

buf = io.BytesIO()
w = wave.open(buf, "wb")
w.setnchannels(2)
w.setsampwidth(2)
w.setframerate(8000)
w.setcomptype("NONE", "not compressed")
frames = b"".join(struct.pack("<hh", i, -i) for i in range(100))
w.writeframes(frames)
print(w.getnframes(), w.getparams())
w.close()
data = buf.getvalue()
print(len(data), data[:4], data[8:12], data[36:40])
r = wave.open(io.BytesIO(data), "rb")
print(r.getnchannels(), r.getsampwidth(), r.getframerate(), r.getnframes(), r.getcomptype(), r.getcompname())
print(r.getparams())
print(r.readframes(3), r.tell())
r.setpos(98)
print(struct.unpack("<4h", r.readframes(5)), r.readframes(1))
r.rewind()
print(r.readframes(1000) == frames)
r.close()
try:
    wave.open(io.BytesIO(b"not a wave file at all"), "rb")
except wave.Error as e:
    print("Error:", e)
try:
    w = wave.open(io.BytesIO(), "wb")
    w.writeframes(b"x")
except wave.Error as e:
    print("Error:", e)
