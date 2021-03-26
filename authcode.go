/*对https://github.com/starten/go-authcode进行river-node适配封装*/

package river_node

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)



const AUTHCODE_RIVERNODE_NAME = "authcode"


type AuthCodeConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.ENCODE或DECODE*/
	
	Raws 		      			<-chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	Salt						[]byte
	TimeoutSec 					time.Duration

	News_Encode					chan []byte
	//----------
	News_Decode					chan []byte
	Limit_Decode      			int
}




func (p *AuthCodeConfig)Name()string{
	return AUTHCODE_RIVERNODE_NAME
}



type AuthCode struct{
	config 				*AuthCodeConfig

	countor 			int
	event_run 			Event
	event_fused 		Event

	stop				chan struct{}
}

func (p *AuthCode)Name()string{
	return AUTHCODE_RIVERNODE_NAME
}

func (p *AuthCode)Construct(AuthCodeConfigAbs Config) error{
	if AuthCodeConfigAbs.Name() != AUTHCODE_RIVERNODE_NAME {
		return errors.New("crc river_node init error, config must CRCConfig")
	}


	v := reflect.ValueOf(AuthCodeConfigAbs)
	c := v.Interface().(*AuthCodeConfig)


	if c.UniqueId == "" {
		return errors.New("authcode river_node init error, uniqueId is nil")
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New("authcode river_node init error, Events or Errors or Raws is nil")
	}

	if c.News_Encode != nil || c.News_Decode != nil{
		return errors.New("authcode river-node init error, News_Encode or News_Decode is not nil")
	}


	if c.Mode != ENCODE&&c.Mode != DECODE {
		return errors.New("authcode river-node init error, unknown mode") 
	}

	if c.Salt ==nil{
		return errors.New("authcode river-node init error, Salt is nil") 
	}

	//c.TimeoutSec 可以为0

	
	if c.Mode ==DECODE && (c.Limit_Decode ==0){
		return errors.New("decode mode authcode river-node init error, "+
		             "Limit_Decode is nil") 
	}

	
	p.config = c


	p.event_fused 	  	= NewEvent(AUTHCODE_FUSED, c.UniqueId, "", "")


	modeStr   := ""

	if p.config.Mode == ENCODE{
		modeStr ="authcode为加密模式，将加密后的数据注入News_EnCode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("authcode适配器开始运行，其UniqueId为%s, Mode为:%s",
			   		  p.config.UniqueId, modeStr))

		p.config.News_Encode		= make(chan []byte)

	}else if p.config.Mode == DECODE{
		modeStr ="authcode为解密模式，将解密后的数据注入News_DeCode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("authcode适配器开始运行，其UniqueId为%s, Mode为:%s",
					  p.config.UniqueId, modeStr))


		p.config.News_Decode 		= make(chan []byte)
	}

	p.stop 				= make(chan struct{})

	return nil

}



func (p *AuthCode)Run(){

	p.config.Events <-p.event_run

	switch p.config.Mode{
	case ENCODE:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.encode(raw)
				case <-p.stop:
					return
				}
			}
		}()
	case DECODE:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.decode(raw)
				case <-p.stop:
					return
				}
			}
		}()
	}
}


func (p *AuthCode)ProactiveDestruct(){
	p.config.Events <-NewEvent(AUTHCODE_PROACTIVEDESTRUCT,p.config.UniqueId,"","注意，由于某些原因authcode包主动调用了显式析构方法")
	p.stop<-struct{}{}	
}

//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *AuthCode)reactiveDestruct(){
	p.config.Events <-NewEvent(AUTHCODE_REACTIVEDESTRUCT,p.config.UniqueId,"","authcode包触发了隐式析构方法")

	//析构数据源
	close(p.stop)

	switch p.config.Mode{
	case ENCODE:
		close(p.config.News_Encode)
	case DECODE:
		close(p.config.News_Decode)
	}
}



func NewAuthCode() NodeAbstract {
	return &AuthCode{}
}


func init() {
	Register(AUTHCODE_RIVERNODE_NAME, NewAuthCode)
	logger.Info("预加载完成，AuthCode适配器已预加载至package river_node.RNodes结构内")
}


/*------------以下是所需的功能方法-------------*/

/*示例：
	str 	:= "hello authcode"
	encode 	:= p.authcode.AuthCode(str, "ENCODE", "", 0)
	fmt.Println(encode)

	decode 	:= p.authcode.AuthCode(encode, "DECODE", "", 0)
	fmt.Println(decode)
*/

func (p *AuthCode)encode(raw []byte){
	if en := p.authcode.AuthCode(string(raw), ENCODE, p.config.Salt, p.config.TimeoutSec);en ==nil{
		return
	}else{
		p.Config.News_Encode <- en
	} 
}

func (p *AuthCode)decode(raw []byte){
	if de := p.authcode.AuthCode(string(raw), DECODE, p.config.Salt, p.config.TimeoutSec);de ==nil{
		return
	}else{
		p.Config.News_Decode <- de
	} 
}

/*------------以下是所需的底层函数-------------*/

// func (enc *Encoding) EncodeToString(src []byte) string {
// 	buf := make([]byte, enc.EncodedLen(len(src)))
// 	enc.Encode(buf, src)
// 	return string(buf)
// }

// base64 加密
func (p *AuthCode)base64_encode(s string) []byte {
	buf := make([]byte, base64.EncodedLen(len(s)))
	base64.Encode(buf,s)
	return buf
}

// func (p *AuthCode)base64_encode(s string) string {
// 	return base64.StdEncoding.EncodeToString([]byte(s))
// }


// func (enc *Encoding) DecodeString(s string) ([]byte, error) {
// 	dbuf := make([]byte, enc.DecodedLen(len(s)))
// 	n, err := enc.Decode(dbuf, []byte(s))
// 	return dbuf[:n], err
// }

//  base64 解密
func (p *AuthCode)base64_decode(s string) ([]byte,error) {
	dbuf := make([]byte, base64.DecodedLen(len(s)))
	n, err := base64.Decode(dbuf, []byte(s))
	return dbuf[:n],err 
}

// func (p *AuthCode)base64_decode(s string) (string,error) {
// 	sByte, err := base64.StdEncoding.DecodeString(s)
// 	if err == nil {
// 		return string(sByte),nil
// 	} else {
// 		return "",err
// 	}
// }

//生成32位md5字串
func (p *AuthCode)getMd5Bytes(b []byte) []byte {
	h := md5.New()//h代表hash
	h.Write(b)
	return h.Sum(nil)
	/**
	 通过翻阅源码可以看到他并不是对data进行校验计算
	 而是对hash.Hash对象内部存储的内容进行校验和
	 计算然后将其追加到data的后面形成一个新的byte切片
	 因此通常的使用方法就是将data置为nil
	 */
}

//生成32位md5字串
func (p *AuthCode)getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
	/**
	 通过翻阅源码可以看到他并不是对data进行校验计算
	 而是对hash.Hash对象内部存储的内容进行校验和
	 计算然后将其追加到data的后面形成一个新的byte切片
	 因此通常的使用方法就是将data置为nil
	 */
}

/**
 * $string 明文或密文
 * $operation 加密ENCODE或解密DECODE
 * $key 密钥
 * $expiry 密钥有效期
 */ 
func (p *AuthCode)authCode_WEB(str string, operation int, key string, expiry int) string {
func (p *AuthCode)authCode(str string, operation int, key string, expiry int) string {
	// 动态密匙长度，相同的明文会生成不同密文就是依靠动态密匙
	// 加入随机密钥，可以令密文无任何规律，即便是原文和密钥完全相同，加密结果也会每次不同，增大破解难度。
	// 取值越大，密文变动规律越大，密文变化 = 16 的 ckey_length 次方
	// 当此值为 0 时，则不产生随机密钥
	ckey_length := 4;
  
	// 密匙禁止为空
	key = p.getMd5String(key)//md5(key)的长度必然为32，key自身的长度倒是无所谓了，将key直接转换成了md5形式
	
	// 密匙a会参与加解密
	keya := p.getMd5String(key[:16])//正好一半
	// 密匙b会用来做数据完整性验证
	keyb := p.getMd5String(key[16:])//正好另一半
	// 密匙c用于变化生成的密文
	keyc := ""
	if ckey_length != 0 {
		if operation == DECODE {
			keyc = str[:ckey_length]//这里用到了str，并且是原始的未加工的str字符串
		} else {
			sTime := p.getMd5String(time.Now().String())//这里拿到的不是时间字符串，而是时间字符串的md5，因此必然会是32位长度
			sLen := 32 - ckey_length
			keyc = sTime[sLen:]//keyc其实就是一个时间戳的md5码片段
		}
	}

	//crypt
	//n.	(尤指旧时作墓穴用的)教堂地下室;

	// 参与运算的密匙
	cryptkey := fmt.Sprintf("%s%s", keya, p.getMd5String(keya + keyc))
	key_length := len(cryptkey)



	//下面才会对具体数据进行操作，也就是str


	// 明文，前10位用来保存时间戳，解密时验证数据有效性，10到26位用来保存$keyb(密匙b)，解密时会通过这个密匙验证数据完整性
	// 如果是解码的话，会从第$ckey_length位开始，因为密文前$ckey_length位保存 动态密匙，以保证解密正确
	if operation == DECODE {
		//通过URL传值取回时一些字符会被转义，所以如果是http协议的应用场景就需要转回来，非http协议长久就不需要转
		//目前的我来说并不需要转化，即使转化也不该在代码主逻辑代码段实现
		str = strings.Replace(str, "*", "+", -1)
		str = strings.Replace(str, "_", "/", -1)
//
		// sByte, err := base64.StdEncoding.DecodeString(s)
		// if err == nil {
		// 	return string(sByte)
		// } else {
		// 	p.Errors <- NewError(AUTHCODE_BASE64_DECODE_FAIL, p.config.UniqueId, s,
		// 			 fmt.Sprintf("authcode适配器在执行base64_decode方法时发生错误，错误内容为:%s,date字段会返回被操作的原始string字符串",err))
		// 	return ""
		// }
//
		if str,err :=p.base64_decode(str[ckey_length:]); err !=nil{
			p.Errors <- NewError(AUTHCODE_BASE64_DECODE_FAIL, p.config.UniqueId, s,
					  fmt.Sprintf("authcode适配器在执行base64_decode方法时发生错误，错误内容为:%s,date字段会返回被操作的原始string字符串",err))
			return ""
		}
		// strByte, err := base64.StdEncoding.DecodeString()
		// if err != nil {
		// 	p.Errors <- NewError(err)
		// 	return ""	
		// }
		// str = string(strByte)
//
	} else {
		if expiry != 0 {expiry = expiry + time.Now().Unix()}
		tmpMd5 := getMd5String(str + keyb)
		str = fmt.Sprintf("%010d%s%s", expiry, tmpMd5[:16], str)//010d不是字符串，而是%010d,代表了十进制整数输出,同时指定了宽度是10位(expiry是个10进制整型变量)
	}
	string_length := len(str)
	resdata := make([]byte, 0, string_length)
	var rndkey, box [256]int
	// 产生密匙簿
	j := 0
	a := 0
	i := 0
	tmp := 0
	for i = 0; i < 256; i++ {
		rndkey[i] = int(cryptkey[i % key_length])
		box[i] = i
	}
	// 用固定的算法，打乱密匙簿，增加随机性，好像很复杂，实际上并不会增加密文的强度
	for i = 0; i < 256; i ++ {
		j = (j + box[i] + rndkey[i]) % 256
		tmp = box[i]
		box[i] = box[j]
		box[j] = tmp
	}
	// 核心加解密部分
	a = 0;j = 0;tmp = 0
	for i = 0; i < string_length; i++ {
		a = ((a + 1) % 256)
		j = ((j + box[a]) % 256)
		tmp = box[a]
		box[a] = box[j]
		box[j] = tmp
		// 从密匙簿得出密匙进行异或，再转成字符
		resdata = append(resdata, byte(int(str[i]) ^ box[(box[a] + box[j]) % 256]))
	}
	result := string(resdata)	
	if operation == DECODE {
		// substr($result, 0, 10) == 0 验证数据有效性
		// substr($result, 0, 10) - time() > 0 验证数据有效性
		// substr($result, 10, 16) == substr(md5(substr($result, 26).$keyb), 0, 16) 验证数据完整性
		// 验证数据有效性，请看未加密明文的格式
		frontTen, _ := strconv.ParseInt(result[:10], 10, 0)
		if (frontTen == 0 || frontTen - time.Now().Unix() > 0) && result[10:26] == getMd5String(result[26:]+keyb)[:16] {
			return result[26:]
		} else {
			p.Errors <- NewError(AUTHCODE_DECODE_FAIL, p.config.UniqueId, result,
					 fmt.Sprintf("authcode适配器在进行解码操作流程的最后一步时发生错误(FAIL),date字段会返回被解码的临时字段“result”的原始string字符串"))
			return "";
		}
	} else {
		// 把动态密匙保存在密文里，这也是为什么同样的明文，生产不同密文后能解密的原因
		// 因为加密后的密文可能是一些特殊字符，复制过程可能会丢失，所以用base64编码
		result = keyc + base64.StdEncoding.EncodeToString([]byte(result))
		result = strings.Replace(result, "+", "*", -1)
		result = strings.Replace(result, "/", "_", -1)

		if result ==""{
			p.Errors <- NewError(AUTHCODE_DECODE_NULL, p.config.UniqueId, result,
					 fmt.Sprintf("authcode适配器在进行解码操作流程的最后一步时发生错误(NULL),date字段会返回被解码的临时字段“result”的原始string字符串"))
			return ""
		}else{
			return result
		}
	}
}