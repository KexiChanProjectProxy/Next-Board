import QRCode from 'qrcode';

/**
 * Generate QR code as data URL
 */
export async function generateQRCode(text: string): Promise<string> {
  try {
    return await QRCode.toDataURL(text, {
      width: 300,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF',
      },
    });
  } catch (error) {
    console.error('Error generating QR code:', error);
    throw error;
  }
}

/**
 * Generate node configuration string for QR code
 */
export function generateNodeConfig(node: any): string {
  // This is a placeholder - actual format depends on node type
  // For example, vmess://... or vless://... or trojan://...
  const config = {
    name: node.name,
    type: node.node_type,
    host: node.host,
    port: node.port,
    ...node.protocol_config,
  };

  // Return base64 encoded JSON for now
  // In production, this should generate proper protocol-specific URLs
  return `${node.node_type}://${btoa(JSON.stringify(config))}`;
}
