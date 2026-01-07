import CryptoJS from 'crypto-js';

const defaultKey = 'f080a463654b2279';

export const aesEcb = {
  // 加密
  encrypt(word: string, keyStr: string = defaultKey): string {
    const key = CryptoJS.enc.Utf8.parse(keyStr);
    const src = CryptoJS.enc.Utf8.parse(word);
    const encrypted = CryptoJS.AES.encrypt(src, key, {
      mode: CryptoJS.mode.ECB,
      padding: CryptoJS.pad.Pkcs7,
    });
    // 返回纯Base64编码的密文，不包含OpenSSL格式头
    return encrypted.ciphertext.toString(CryptoJS.enc.Base64);
  },
  // 解密
  decrypt(word: string, keyStr: string = defaultKey): string {
    const key = CryptoJS.enc.Utf8.parse(keyStr);
    // 将Base64字符串转换为CipherParams对象
    const cipherParams = CryptoJS.lib.CipherParams.create({
      ciphertext: CryptoJS.enc.Base64.parse(word),
    });
    const decrypt = CryptoJS.AES.decrypt(cipherParams, key, {
      mode: CryptoJS.mode.ECB,
      padding: CryptoJS.pad.Pkcs7,
    });
    return CryptoJS.enc.Utf8.stringify(decrypt).toString();
  },
};
