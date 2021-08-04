/**
 * Autogenerated by Thrift Compiler (0.12.0)
 *
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */
#ifndef TextService_H
#define TextService_H

#include <thrift/TDispatchProcessor.h>
#include <thrift/async/TConcurrentClientSyncInfo.h>
#include "media_service_types.h"

namespace media_service {

#ifdef _MSC_VER
  #pragma warning( push )
  #pragma warning (disable : 4250 ) //inheriting methods via dominance 
#endif

class TextServiceIf {
 public:
  virtual ~TextServiceIf() {}
  virtual void UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier) = 0;
};

class TextServiceIfFactory {
 public:
  typedef TextServiceIf Handler;

  virtual ~TextServiceIfFactory() {}

  virtual TextServiceIf* getHandler(const ::apache::thrift::TConnectionInfo& connInfo) = 0;
  virtual void releaseHandler(TextServiceIf* /* handler */) = 0;
};

class TextServiceIfSingletonFactory : virtual public TextServiceIfFactory {
 public:
  TextServiceIfSingletonFactory(const ::apache::thrift::stdcxx::shared_ptr<TextServiceIf>& iface) : iface_(iface) {}
  virtual ~TextServiceIfSingletonFactory() {}

  virtual TextServiceIf* getHandler(const ::apache::thrift::TConnectionInfo&) {
    return iface_.get();
  }
  virtual void releaseHandler(TextServiceIf* /* handler */) {}

 protected:
  ::apache::thrift::stdcxx::shared_ptr<TextServiceIf> iface_;
};

class TextServiceNull : virtual public TextServiceIf {
 public:
  virtual ~TextServiceNull() {}
  void UploadText(const int64_t /* req_id */, const std::string& /* text */, const std::map<std::string, std::string> & /* carrier */) {
    return;
  }
};

typedef struct _TextService_UploadText_args__isset {
  _TextService_UploadText_args__isset() : req_id(false), text(false), carrier(false) {}
  bool req_id :1;
  bool text :1;
  bool carrier :1;
} _TextService_UploadText_args__isset;

class TextService_UploadText_args {
 public:

  TextService_UploadText_args(const TextService_UploadText_args&);
  TextService_UploadText_args& operator=(const TextService_UploadText_args&);
  TextService_UploadText_args() : req_id(0), text() {
  }

  virtual ~TextService_UploadText_args() throw();
  int64_t req_id;
  std::string text;
  std::map<std::string, std::string>  carrier;

  _TextService_UploadText_args__isset __isset;

  void __set_req_id(const int64_t val);

  void __set_text(const std::string& val);

  void __set_carrier(const std::map<std::string, std::string> & val);

  bool operator == (const TextService_UploadText_args & rhs) const
  {
    if (!(req_id == rhs.req_id))
      return false;
    if (!(text == rhs.text))
      return false;
    if (!(carrier == rhs.carrier))
      return false;
    return true;
  }
  bool operator != (const TextService_UploadText_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const TextService_UploadText_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class TextService_UploadText_pargs {
 public:


  virtual ~TextService_UploadText_pargs() throw();
  const int64_t* req_id;
  const std::string* text;
  const std::map<std::string, std::string> * carrier;

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _TextService_UploadText_result__isset {
  _TextService_UploadText_result__isset() : se(false) {}
  bool se :1;
} _TextService_UploadText_result__isset;

class TextService_UploadText_result {
 public:

  TextService_UploadText_result(const TextService_UploadText_result&);
  TextService_UploadText_result& operator=(const TextService_UploadText_result&);
  TextService_UploadText_result() {
  }

  virtual ~TextService_UploadText_result() throw();
  ServiceException se;

  _TextService_UploadText_result__isset __isset;

  void __set_se(const ServiceException& val);

  bool operator == (const TextService_UploadText_result & rhs) const
  {
    if (!(se == rhs.se))
      return false;
    return true;
  }
  bool operator != (const TextService_UploadText_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const TextService_UploadText_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _TextService_UploadText_presult__isset {
  _TextService_UploadText_presult__isset() : se(false) {}
  bool se :1;
} _TextService_UploadText_presult__isset;

class TextService_UploadText_presult {
 public:


  virtual ~TextService_UploadText_presult() throw();
  ServiceException se;

  _TextService_UploadText_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};

class TextServiceClient : virtual public TextServiceIf {
 public:
  TextServiceClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  TextServiceClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier);
  void send_UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier);
  void recv_UploadText();
 protected:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
};

class TextServiceProcessor : public ::apache::thrift::TDispatchProcessor {
 protected:
  ::apache::thrift::stdcxx::shared_ptr<TextServiceIf> iface_;
  virtual bool dispatchCall(::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, const std::string& fname, int32_t seqid, void* callContext);
 private:
  typedef  void (TextServiceProcessor::*ProcessFunction)(int32_t, ::apache::thrift::protocol::TProtocol*, ::apache::thrift::protocol::TProtocol*, void*);
  typedef std::map<std::string, ProcessFunction> ProcessMap;
  ProcessMap processMap_;
  void process_UploadText(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
 public:
  TextServiceProcessor(::apache::thrift::stdcxx::shared_ptr<TextServiceIf> iface) :
    iface_(iface) {
    processMap_["UploadText"] = &TextServiceProcessor::process_UploadText;
  }

  virtual ~TextServiceProcessor() {}
};

class TextServiceProcessorFactory : public ::apache::thrift::TProcessorFactory {
 public:
  TextServiceProcessorFactory(const ::apache::thrift::stdcxx::shared_ptr< TextServiceIfFactory >& handlerFactory) :
      handlerFactory_(handlerFactory) {}

  ::apache::thrift::stdcxx::shared_ptr< ::apache::thrift::TProcessor > getProcessor(const ::apache::thrift::TConnectionInfo& connInfo);

 protected:
  ::apache::thrift::stdcxx::shared_ptr< TextServiceIfFactory > handlerFactory_;
};

class TextServiceMultiface : virtual public TextServiceIf {
 public:
  TextServiceMultiface(std::vector<apache::thrift::stdcxx::shared_ptr<TextServiceIf> >& ifaces) : ifaces_(ifaces) {
  }
  virtual ~TextServiceMultiface() {}
 protected:
  std::vector<apache::thrift::stdcxx::shared_ptr<TextServiceIf> > ifaces_;
  TextServiceMultiface() {}
  void add(::apache::thrift::stdcxx::shared_ptr<TextServiceIf> iface) {
    ifaces_.push_back(iface);
  }
 public:
  void UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->UploadText(req_id, text, carrier);
    }
    ifaces_[i]->UploadText(req_id, text, carrier);
  }

};

// The 'concurrent' client is a thread safe client that correctly handles
// out of order responses.  It is slower than the regular client, so should
// only be used when you need to share a connection among multiple threads
class TextServiceConcurrentClient : virtual public TextServiceIf {
 public:
  TextServiceConcurrentClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  TextServiceConcurrentClient(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier);
  int32_t send_UploadText(const int64_t req_id, const std::string& text, const std::map<std::string, std::string> & carrier);
  void recv_UploadText(const int32_t seqid);
 protected:
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  apache::thrift::stdcxx::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
  ::apache::thrift::async::TConcurrentClientSyncInfo sync_;
};

#ifdef _MSC_VER
  #pragma warning( pop )
#endif

} // namespace

#endif
