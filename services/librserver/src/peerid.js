// libp2p peer IDs for Ed25519 keys: identity multihash of the protobuf PublicKey
// {Type: Ed25519, Data: key}, base58btc-encoded ("12D3KooW...").
const ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';

function base58(buf) {
  let n = BigInt('0x' + (buf.toString('hex') || '0'));
  let out = '';
  while (n > 0n) {
    out = ALPHABET[Number(n % 58n)] + out;
    n /= 58n;
  }
  for (const b of buf) {
    if (b !== 0) break;
    out = '1' + out;
  }
  return out;
}

function peerIdFromEd25519(raw) {
  const pb = Buffer.concat([Buffer.from([0x08, 0x01, 0x12, 0x20]), raw]);
  return base58(Buffer.concat([Buffer.from([0x00, pb.length]), pb]));
}

module.exports = { peerIdFromEd25519 };
