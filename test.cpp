/************************************************************************CPY11*/
/*   Copyright Mentor Graphics Corporation 2013  All Rights Reserved.    CPY12*/
/*                                                                       CPY13*/
/*   THIS WORK CONTAINS TRADE SECRET AND PROPRIETARY INFORMATION         CPY14*/
/*   WHICH IS THE PROPERTY OF MENTOR GRAPHICS CORPORATION OR ITS         CPY15*/
/*   LICENSORS AND IS SUBJECT TO LICENSE TERMS.                          CPY16*/
/************************************************************************CPY17*/
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
//
//  MODULE:        hen_utils.cpp
//
//  AUTHOR:        Levon Manukyan
//
//  DESCRIPTION:
//
//  NOTES:
//
//
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

#include <ponte/hen/hen_utils.hpp>
#include <ponte/hen/predefinedproperties.hpp>
#include <ponte/hen/cellinsttransform.hpp>

#include <ponte/containers/fifo.hpp>

#include <ponte/base/typeswitchtable.hpp>
#include <ponte/base/stdutility.hpp>
#include <ponte/base/errorhandling.hpp>
#include <ponte/base/safe_static_cast.hpp>

#include <ponte/iterators/foreach.hpp>

namespace HEN
{

using namespace PredefinedProps;

GetCellInstTransform::CellInstProps::CellInstProps(const PropMgr* propMgr)
    : x(propMgr, CELLINST_X_PROP)
    , y(propMgr, CELLINST_Y_PROP)
    , orientation(propMgr, CELLINST_ORIENTATION_PROP)
{}

GetCellInstTransform::GetCellInstTransform(const IHEN* hen)
    : cellInstProps_(&hen->cellInstProperties())
{}

GeometricalTransform
GetCellInstTransform::getTransform(const FQCellInstanceRef& subCell) const
{
    typedef Ponte::Orientation::Type Orientation;
    const CoordPropType& x = cellInstProps_.x.getPropertyValue(subCell);
    const CoordPropType& y = cellInstProps_.y.getPropertyValue(subCell);
    const OrientationPropType o = cellInstProps_.orientation.getPropertyValue(subCell);

    return GeometricalTransform(x, y, static_cast<Orientation>(o));
}

CellInstTransform::CellInstProps::CellInstProps(PropMgr* propMgr)
    : x(propMgr, CELLINST_X_PROP)
    , y(propMgr, CELLINST_Y_PROP)
    , orientation(propMgr, CELLINST_ORIENTATION_PROP)
{}

CellInstTransform::CellInstTransform(IHEN* hen)
    : cellInstProps_(&hen->cellInstProperties())
{}

void
CellInstTransform::Proxy::operator=(const GeometricalTransform& t) const
{
    cellInstProps_.x.setPropertyValue(subCell_, t.xTranslation());
    cellInstProps_.y.setPropertyValue(subCell_, t.yTranslation());
    cellInstProps_.orientation.setPropertyValue(subCell_, t.orientation().type());
}

Utils::HierPathObtainer::HierPathObtainer(const IHEN& hen)
    : cellInstNamePropRef_(hen.cellInstProperties().getPropertyRef<std::string>(PredefinedProps::CELLINST_NAME_PROP))
{}

std::string
Utils::HierPathObtainer::getHierPath(const CellInstData& cellInstData) const
{
    if ( cellInstData.depth() == 0 )
        return std::string();

    return cellInstData.getValue(cellInstNamePropRef_);
}

////////////////////// NetMapper ////////////////////////
Utils::NetMapper::NetMapper()
{
}

void
Utils::NetMapper::mapNodeRef(ulong nodeNumber, NodeRef nodeRef)
{
    PONTE_ASSERT(nodeNumber == unwrapping_cast(nodeRef) + 1, "");
}

NodeRef
Utils::NetMapper::getNodeRef(ulong nodeNumber) const
{
    PONTE_ASSERT(nodeNumber > 0, "");
    return NodeRef(nodeNumber - 1);
}

ulong
Utils::NetMapper::getNodeNumber(NodeRef nodeRef) const
{
    return unwrapping_cast(nodeRef) + 1;
}

void
Utils::NetMapper::mapPin(ulong pinNumber, const NodeRef& pinNodeRef)
{
    PONTE_ASSERT(pinNodeRef != IHEN::invalidNodeRef(), "");
    PONTE_ASSERT(!contains(pinMap_, pinNumber), "");
    pinMap_[pinNumber] = pinNodeRef;
}

NodeRef
Utils::NetMapper::getPinNodeRef(ulong pinNumber) const
{
    const PinMap::const_iterator it = pinMap_.find(pinNumber);
    //if(it == pinMap_.end()) return IHEN::invalidNodeRef();
    PONTE_ASSERT(it != pinMap_.end(), "");
    return it->second;
}

////////////////////// CellMapper ///////////////////////
Utils::CellMapper::CellMapper()
{
}

void
Utils::CellMapper::clearNetsMap()
{
    netMappers_.clear();
    netRefMap_.clear();
}

void
Utils::CellMapper::clear()
{
  clearNetsMap();
}

Utils::NetMapper&
Utils::CellMapper::getNetMapper(ulong n) const
{
  const NetRefMapper::const_iterator it = netMappers_.find(n);
  if(it == netMappers_.end())
      netMappers_[n] = boost::shared_ptr<NetMapper>(new NetMapper());
  return *netMappers_[n];
}

void
Utils::CellMapper::mapNetRef(ulong netNumber, NetRef netRef)
{
    PONTE_ASSERT(netRef != IHEN::invalidNetRef(), "");
    PONTE_ASSERT(!contains(netRefMap_, netNumber), "");
    netRefMap_[netNumber] = netRef;
}

NetRef
Utils::CellMapper::getNetRef(ulong netNumber) const
{
    const NetRefMap::const_iterator it = netRefMap_.find(netNumber);
    //if(it == netRefMap_.end()) return IHEN::invalidNetRef();
    PONTE_ASSERT(it != netRefMap_.end(), "");
    return it->second;
}

void
Utils::CellMapper::mapCellInstRef(ulong placementNumber, CellInstanceRef cellInstRef)
{
    PONTE_ASSERT(cellInstRef != IHEN::invalidCellInstanceRef(), "");
    PONTE_ASSERT(!contains(cellInstRefMap_, placementNumber), "");
    cellInstRefMap_[placementNumber] = cellInstRef;
}

CellInstanceRef
Utils::CellMapper::getCellInstRef(ulong placementNumber) const
{
    const CellInstRefMap::const_iterator it = cellInstRefMap_.find(placementNumber);
    if(it == cellInstRefMap_.end())
        return IHEN::invalidCellInstanceRef();
    return it->second;
}

void
Utils::CellMapper::mapDevInstRef(ulong devInstId, CellInstanceRef devInstRef)
{
    PONTE_ASSERT(!contains(devInstRefMap_, devInstId), "");
    devInstRefMap_[devInstId] = devInstRef;
}

CellInstanceRef
Utils::CellMapper::getDevInstRef(ulong devInstId) const
{
    const DevInstRefMap::const_iterator it = devInstRefMap_.find(devInstId);
    if ( it == devInstRefMap_.end() )
        return IHEN::invalidCellInstanceRef();
    return it->second;
}

namespace
{

class DeepSubCellDataPromoter
{
    typedef IHEN::CellInstProperties CellInstPropMgr;

    class ISubCellPropertyValuePromoter
    {
    public:
        virtual ~ISubCellPropertyValuePromoter() {}
        virtual void setPropertyValue(CellInstData& promotedSubCellData, FQCellInstanceRef subCell) const = 0;
        virtual void promotePropertyValue(const CellInstData& parentData, const CellInstData& subCellData, CellInstData& promotedSubCellData) const = 0;
    };

    class ISubCellPropertyValuePromoterFactory
    {
    public:
        virtual ~ISubCellPropertyValuePromoterFactory() {}
        virtual ISubCellPropertyValuePromoter* create(const IHEN& hen, IPropertyManagerBase::UntypedPropertyRef propRef) const = 0;
    };

    static void addPropertValues(const CellInstData& data, const IHEN::CellInstancePropertyRefCollection& propRefs, HierObjectProperties::PropertyValues& propValues)
    {
        for (size_t i = 0; i < propRefs.size(); ++i)
        {
            propValues.add(data.getPropValuePtr(propRefs[i])->clone());
        }
    }

    template<class T>
    class SubCellPropertyValuePromoterFactory : public ISubCellPropertyValuePromoterFactory
    {
        class SubCellPropertyValuePromoter : public ISubCellPropertyValuePromoter
        {
            const IHEN::CellInstPropertyPromotion::Data* promotionData_;
            const CellInstPropMgr::ConstProperty<T> prop_;


            virtual void setPropertyValue(CellInstData& promotedSubCellData, FQCellInstanceRef subCell) const
            {
                promotedSubCellData.setValue(prop_.propertyRef(), prop_.getPropertyValue(subCell));
            }

            virtual void promotePropertyValue(const CellInstData& parentData, const CellInstData& subCellData, CellInstData& promotedSubCellData) const
            {
                PONTE_TRACE_BLOCK(80, HEN, "SubCellPropertyValuePromoter::promotePropertyValue " << " propertyRef = " << unwrapping_cast(prop_.propertyRef()) << "");
                // TODO: make a member
                HierObjectProperties hop(promotionData_->promoter().inputHierObjectSpecifierDescriptor());
                const IHEN::CellInstPropertyPromotion::Info& promotionInfo = promotionData_->promotionInfo();
                addPropertValues(parentData, promotionInfo.subcellProps, hop.propertyValues(0));
                addPropertValues(subCellData, promotionInfo.subsubcellProps, hop.propertyValues(1));

                const T& promotedPropValue = promotionData_->promoter().template getPromotedValue<T>(hop);
                PONTE_TRACE(80, HEN, "promotedPropValue = " << promotedPropValue );
                promotedSubCellData.setValue(prop_.propertyRef(), promotedPropValue);
            }
        public:
            SubCellPropertyValuePromoter(const IHEN::CellInstPropertyPromotion::Data* promData, const CellInstPropMgr::ConstProperty<T>& prop)
                : promotionData_(promData)
                , prop_(prop)
            {
                if ( promotionData_ == NULL )
                {
                    not_implemented("");
                }
            }
        };

    public:
        virtual ISubCellPropertyValuePromoter* create(const IHEN& hen, IPropertyManagerBase::UntypedPropertyRef propRef) const
        {
            const CellInstPropMgr& cellInstPropMgr = hen.cellInstProperties();
            const CellInstPropMgr::PropertyRef<T> typedPropRef = cellInstPropMgr.template getTypedPropertyRef<T>(propRef);
            const IHEN::CellInstPropertyPromotion::Data* promData = hen.cellInstPropertyPromotionSettings().get(typedPropRef);
            const CellInstPropMgr::ConstProperty<T> prop = cellInstPropMgr.template getProperty<T>(typedPropRef);
            return new SubCellPropertyValuePromoter(promData, prop);
        }
    };

public:
    DeepSubCellDataPromoter(const IHEN& hen)
        : promoters_(hen)
    {
    }

    CellInstData getPromotedData(const CellInstData& parentData, FQCellInstanceRef subCell) const
    {
        CellInstData subCellData(parentData.depth() + 1);
        CellInstPropMgr::PropertyIterator propIt = promoters_.getProperties();
        for ( ; propIt != null(); ++propIt )
            promoters_.getPtr(*propIt)->setPropertyValue(subCellData, subCell);

        if ( parentData.depth() == 0 )
            return subCellData;

        CellInstData promotedSubCellData(subCellData);
        for ( propIt = promoters_.getProperties(); propIt != null(); ++propIt )
            promoters_.getPtr(*propIt)->promotePropertyValue(parentData, subCellData, promotedSubCellData);
        return promotedSubCellData;
    }

private:
    class PromoterTable : public PropertyRefIndexedPtrMap<ISubCellPropertyValuePromoter>
    {
        typedef TypeSwitchTable<ISubCellPropertyValuePromoterFactory, SupportedPropTypes>::FilledWith<SubCellPropertyValuePromoterFactory> PromoterFactory;
        PromoterFactory promoterFactory_;
        const IHEN& hen_;

        ISubCellPropertyValuePromoter* createPromoter(const IHEN& hen, IPropertyManagerBase::UntypedPropertyRef propRef) const
        {
            const IPropertyManagerBase::PropertyDescriptor& pd = cellInstPropMgr().getPropertyDescriptor(propRef);
            return promoterFactory_[strong_cast<unsigned>(pd.propertyTypeId())].create(hen, propRef);
        }

        void init()
        {
            CellInstPropMgr::PropertyIterator propIt = getProperties();
            for ( ; propIt != null(); ++propIt )
                set(*propIt, createPromoter(hen_, *propIt));
        }

    public:
        PromoterTable(const IHEN& hen)
            : hen_(hen)
        {
            init();
        }

        PromoterTable(const PromoterTable& other)
            : promoterFactory_()
            , hen_(other.hen_)
        {
            init();
        }

        CellInstPropMgr::PropertyIterator getProperties() const
        {
            return cellInstPropMgr().getProperties();
        }

        const CellInstPropMgr& cellInstPropMgr() const { return hen_.cellInstProperties(); }
    };


    PromoterTable promoters_;
};

class FlatNetIteratorImpl : public Iterators::FwdItTag<const NetInst&>
{
    typedef IHEN::SubNetRef SubNetRef;

    typedef Utils::IFlatNetFilter IFlatNetFilter;
    typedef IFlatNetFilter::Status SubCellStatus;

    struct QueueEntry : public NetInst
    {
        CellInstanceRef cellInst;

        QueueEntry(FQNetRef netRef, const CellInstData& cellInstData)
            : NetInst(netRef, cellInstData)
            , cellInst(IHEN::invalidCellInstanceRef())
        {}

        QueueEntry(FQCellInstanceRef subCell, const CellInstData& cellInstData)
            : NetInst(FQNetRef(subCell.cell, IHEN::invalidNetRef()), cellInstData)
            , cellInst(subCell.subCell)
        {}

        // QueueEntry with an invalid netRef is considered to be a placeholder,
        // when meeting it in the sequence all the local nets of the cell
        // instance are added to the collection.
        bool isSubCellEntry() const
        {
            return net == IHEN::invalidNetRef();
        }
    };

public:
    FlatNetIteratorImpl(const IHEN& hen, CellRef cell, const IFlatNetFilter* netFilter)
        : hen_(hen)
        , netFilter_(netFilter)
        , subCellDataPromoter_(hen)
    {
        addOwnNets(cell);
        addSubCells(cell, CellInstData());
        expandSubCellEntries();
    }

private:

    void addFlatNet(const FQNetRef& net, const CellInstData& cellInstData)
    {
        const QueueEntry netInfo(net, cellInstData);
        if (netFilter_ == NULL || netFilter_->accept(netInfo))
            flatNetInfoQueue_.push(netInfo);
    }

    // This function adds cell's own nets to the collection
    void addOwnNets(CellRef cell)
    {
        typedef IHEN::NetIterator NetIterator;
        for (NetIterator netIt = hen_.getNets(cell); netIt != null(); ++netIt)
        {
            addFlatNet(FQNetRef(cell, *netIt), CellInstData());
        }
    }

    SubCellStatus subCellStatus(const QueueEntry& subCellInfo) const
    {
        if ( netFilter_ == NULL )
            return SubCellStatus(IFlatNetFilter::ACCEPT | IFlatNetFilter::EXPAND);

        return netFilter_->status(subCellInfo.cellInstData);
    }

    void expandSubCell(const QueueEntry& subCellInfo)
    {
        PONTE_ASSERT(subCellInfo.net == IHEN::invalidNetRef(), "");
        PONTE_ASSERT(subCellInfo.cellInst != IHEN::invalidCellInstanceRef(), "");

        const SubCellStatus thisSubCellStatus = subCellStatus(subCellInfo);
        if ( thisSubCellStatus == IFlatNetFilter::REJECT )
            return;

        const CellRef cell = subCellInfo.cell;
        const CellRef subCell = hen_.getCellInstPrototype(cell/subCellInfo.cellInst);
        if ( (thisSubCellStatus & IFlatNetFilter::ACCEPT) != 0)
        {
            // Add all the local nets of the sub cell
            for ( IHEN::NetIterator netIt = hen_.getNets(subCell); netIt != null(); ++netIt )
            {
                const NetRef net = *netIt;
                const SubNetRef subNet(subCellInfo.cellInst, net);
                if (hen_.getConnectedNet(cell, subNet) == IHEN::invalidNetRef())
                {
                    const FQNetRef fqNet(subCell, net);
                    addFlatNet(fqNet, subCellInfo.cellInstData);
                }
            }
        }

        if ( (thisSubCellStatus & IFlatNetFilter::EXPAND) != 0)
        {
            addSubCells(subCell, subCellInfo.cellInstData);
        }
    }

    // This function adds placeholders for all the subcells of the specified
    // cell
    void addSubCells(CellRef cell, const CellInstData& cellData)
    {
        typedef IHEN::CellInstanceIterator CellInstanceIterator;
        for ( CellInstanceIterator subCellIt = hen_.getSubCells(cell); subCellIt != null(); ++subCellIt )
        {
            const FQCellInstanceRef subCell(cell, *subCellIt);
            const CellInstData subCellData = subCellDataPromoter_.getPromotedData(cellData, subCell);

            flatNetInfoQueue_.push(QueueEntry(subCell, subCellData));
        }
    }

    void expandSubCellEntries()
    {
        while (!flatNetInfoQueue_.empty() && flatNetInfoQueue_.get().isSubCellEntry() )
        {
            const QueueEntry subCellEntry = flatNetInfoQueue_.get();
            flatNetInfoQueue_.pop();
            expandSubCell(subCellEntry);
        }
    }

public:
    void operator++()
    {
        PONTE_ASSERT(!flatNetInfoQueue_.empty(), "");

        flatNetInfoQueue_.pop();
        expandSubCellEntries();
    }

    const NetInst& operator*() const
    {
        PONTE_ASSERT(*this != null(), "");
        PONTE_ASSERT(!flatNetInfoQueue_.get().isSubCellEntry(), "");

        return flatNetInfoQueue_.get();
    }

    bool operator==(Null n) const
    {
        return flatNetInfoQueue_.empty();
    }

private:
    typedef FIFO<QueueEntry> Queue;
    const IHEN& hen_;
    Queue flatNetInfoQueue_;
    const IFlatNetFilter* netFilter_;
    DeepSubCellDataPromoter subCellDataPromoter_;
};

template<class RefType>
class PromotedPropertyObtainerBase : public IPropertyObtainer<RefType>
{
protected:
    typedef PropertyMgrCfg<RefType, SupportedPropTypes> Cfg;
    typedef IReadOnlyPropertyAccessor<Cfg> PropAccessorInterfaceType;
    typedef IPropertyBase<RefType> IPropertyType;
    typedef IPropertyManagerBase::PropertyDescriptor PropertyDescriptor;
    typedef IPropertyManagerBase::UntypedPropertyRef UntypedPropertyRef;
    typedef typename PropAccessorInterfaceType::template PropertyRef<UnknownType> GraphItemPropRef;

    typedef GraphItemRelatedTypeResolver<RefType> GraphItemRelatedTypes;
    typedef typename GraphItemRelatedTypes::PropertyPromotionInfoType PropertyPromotionInfoType;
    typedef typename GraphItemRelatedTypes::PromotionDataType PromotionDataType;
    typedef typename GraphItemRelatedTypes::GraphItemPropRefCollection GraphItemPropRefCollection;

    const IHEN& hen_;
    const PropAccessorInterfaceType& propAccessor_;
    const CellInstData& cellInstPropValues_;


    class GraphItemPropValues : public HierObjectProperties
    {
        typedef typename HierObjectProperties::PropertyValues PropertyValues;
    public:
        explicit GraphItemPropValues(const HierObjectSpecifierDescriptor& hierObjSpec)
            : HierObjectProperties(hierObjSpec.specifierLength())
        {
            PONTE_ASSERT(hierObjSpec.specifierLength() == 2, "");
            static_cast<HierObjectSpecifierDescriptor&>(*this) = hierObjSpec;
        }

        GraphItemPropValues(const GraphItemPropValues& other)
            : HierObjectProperties(other.specifierLength())
        {
            PONTE_ASSERT(other.specifierLength() == 2, "");
            static_cast<HierObjectSpecifierDescriptor&>(*this) = other;

            for ( size_t i = 0; i < this->specifierLength(); ++i )
            {
                PropertyValues& pv = propertyValues(i);
                const PropertyValues& opv = other.propertyValues(i);
                for ( size_t j = 0; j < opv.size(); ++j )
                    pv.add(opv[j]->clone());
            }
        }

        mutable RefType currentItemRef;
        void addCellInstPropValue(IValue* propValue)
        {
            HierObjectProperties::propertyValues(0).add(propValue);
        }

        void addGraphItemPropValue(IValue* propValue)
        {
            HierObjectProperties::propertyValues(1).add(propValue);
        }
    };

    class PromotablePropertyBase : public IPropertyType
    {
    protected:
        const IHierObjectPropertyValuePromoterBase& promoter_;
    public:
        explicit PromotablePropertyBase(const IHierObjectPropertyValuePromoterBase& promoter)
            : promoter_(promoter)
            , graphItemPropValues(promoter.inputHierObjectSpecifierDescriptor())
        {}

        GraphItemPropValues graphItemPropValues;
    };

    class IFactory
    {
    public:
        virtual ~IFactory() {}

        virtual IValue* createGraphItemPropertyValue(const RefType* ref, std::auto_ptr<const IPropertyType> prop) const = 0;
        virtual PromotablePropertyBase* createProperty(const PromotionDataType& promData) const = 0;
    };

    template<class T>
    class Factory : public IFactory
    {
        class GraphItemPropValue : public ITypedValue<T>
        {
            const RefType* currentRef_;
            const ::boost::shared_ptr<const IPropertyType> prop_;

            virtual const std::string& nameImpl() const { must_never_reach_here(""); void* p = NULL; return *static_cast<const std::string*>(p); }
            virtual T getValueImpl() const
            {
                T t;
                prop_->typeUnsafeGet(*currentRef_, &t);
                return t;
            }

            virtual void setValueImpl(const T& value)
            {
                must_never_reach_here("");
            }

            virtual IValue* cloneImpl() const
            {
                return new GraphItemPropValue(*this);
            }

        public:
            GraphItemPropValue(const RefType* refPtr, std::auto_ptr<const IPropertyType> prop)
                : currentRef_(refPtr)
                , prop_(prop.release())
            {}

            void setCurrentRefPtr(const RefType* refPtr) { currentRef_ = refPtr; }
        };

        class PromotablePropertyImpl : public PromotablePropertyBase
        {
        public:
            PromotablePropertyImpl(const PromotablePropertyImpl& other)
                : PromotablePropertyBase(other)
            {
                HierObjectProperties::PropertyValues& pv = this->graphItemPropValues.propertyValues(1);
                for(size_t i = 0; i < pv.size(); ++i)
                {
                    GraphItemPropValue& gipv = safe_static_cast<GraphItemPropValue&>(*pv[i]);
                    gipv.setCurrentRefPtr(&this->graphItemPropValues.currentItemRef);
                }
            }

            explicit PromotablePropertyImpl(const IHierObjectPropertyValuePromoterBase& promoter)
                : PromotablePropertyBase(promoter)
            {}

            void get(const RefType& itemRef, T& result) const
            {
                this->graphItemPropValues.currentItemRef = itemRef;
                result = this->promoter_.template getPromotedValue<T>(this->graphItemPropValues);
            }
        };

        typedef ReadOnlyPropertyImpl<T, PromotablePropertyImpl> PromotableProperty;

        virtual IValue* createGraphItemPropertyValue(const RefType* ref, std::auto_ptr<const IPropertyType> prop) const
        {
            return new GraphItemPropValue(ref, prop);
        }

        virtual PromotableProperty* createProperty(const PromotionDataType& promData) const
        {
            return new PromotableProperty(promData.promoter());
        }
    };

    typedef TypeSwitchTable<IFactory, SupportedPropTypes> FactoryTable;
    static const IFactory& factory(const PropertyDescriptor& pd)
    {
        static const typename FactoryTable::template FilledWith<Factory> factoryTable;
        return factoryTable[unwrapping_cast(pd.propertyTypeId())];
    }

protected:
    explicit PromotedPropertyObtainerBase(const IHEN& hen, const CellInstData& cellInstPropValues, const PropAccessorInterfaceType& propAccessor)
        : hen_(hen)
        , propAccessor_(propAccessor)
        , cellInstPropValues_(cellInstPropValues)
    {
    }

    virtual std::auto_ptr<const IPropertyType> get(UntypedPropertyRef propRef) const
    {
        if ( cellInstPropValues_.depth() > 0 )
        {
            const PropertyDescriptor& pd = propAccessor_.getPropertyDescriptor(propRef);
            const PromotionDataType* promData = getPromotionData(propRef);
            if ( promData )
            {
                PromotablePropertyBase* prop = factory(pd).createProperty(*promData);

                const IHEN::CellInstancePropertyRefCollection& cellInstPropRefs = promData->promotionInfo().subcellProps;
                for (size_t i = 0; i < cellInstPropRefs.size(); ++i)
                    prop->graphItemPropValues.addCellInstPropValue(cellInstPropValues_.getPropValuePtr(cellInstPropRefs[i])->clone());

                const RefType* const refPtr = &prop->graphItemPropValues.currentItemRef;
                const GraphItemPropRefCollection& graphItemPropRefs = GraphItemRelatedTypes::getGraphItemPropRefs(promData->promotionInfo());
                for (size_t i = 0; i < graphItemPropRefs.size(); ++i)
                {
                    const GraphItemPropRef propRef1 = graphItemPropRefs[i];
                    const PropertyDescriptor& pd1 = propAccessor_.getPropertyDescriptor(propRef1);
                    std::auto_ptr<const IPropertyType> prop1(propAccessor_.getProperty(propRef1).readonlyProp().clone());
                    IValue* graphItemPropValue = factory(pd1).createGraphItemPropertyValue(refPtr, prop1);
                    prop->graphItemPropValues.addGraphItemPropValue(graphItemPropValue);
                }
                return std::auto_ptr<const IPropertyType>(prop);
            }
        }

        const GraphItemPropRef propRef1 = propAccessor_.template getTypedPropertyRef<UnknownType>(propRef);
        const typename PropAccessorInterfaceType::template ConstProperty<UnknownType> prop = propAccessor_.getProperty(propRef1);
        std::auto_ptr<const IPropertyType> ret(prop.readonlyProp().clone());
        return ret;

    }

    virtual const PromotionDataType* getPromotionData(UntypedPropertyRef propRef) const = 0;
};

class PromotedVertexPropertyObtainer : public PromotedPropertyObtainerBase<NodeRef>
{
    typedef PromotedPropertyObtainerBase<NodeRef> BaseType;
    typedef BaseType::PropAccessorInterfaceType PropAccessorInterfaceType;

public:
    explicit PromotedVertexPropertyObtainer(const IHEN& hen, const CellInstData& cellInstPropValues, const PropAccessorInterfaceType& propAccessor)
        : BaseType(hen, cellInstPropValues, propAccessor)
    {
    }

    virtual const BaseType::PromotionDataType* getPromotionData(UntypedPropertyRef propRef) const
    {
        const IHEN::GraphVertexPropertyManager& graphVertexPropMgr = hen_.getGraphVertexPropertyManager();
        return hen_.getGraphVertexPropertyValuePromotionData(graphVertexPropMgr.getTypedPropertyRef<UnknownType>(propRef));
    }
};

class PromotedEdgePropertyObtainer : public PromotedPropertyObtainerBase<EdgeRef>
{
    typedef PromotedPropertyObtainerBase<EdgeRef> BaseType;
    typedef BaseType::PropAccessorInterfaceType PropAccessorInterfaceType;

public:
    explicit PromotedEdgePropertyObtainer(const IHEN& hen, const CellInstData& cellInstPropValues, const PropAccessorInterfaceType& propAccessor)
        : BaseType(hen, cellInstPropValues, propAccessor)
    {
    }

    virtual const BaseType::PromotionDataType* getPromotionData(UntypedPropertyRef propRef) const
    {
        const IHEN::GraphEdgePropertyManager& graphEdgePropMgr = hen_.getGraphEdgePropertyManager();
        return hen_.getGraphEdgePropertyValuePromotionData(graphEdgePropMgr.getTypedPropertyRef<UnknownType>(propRef));
    }
};

FQCellInstanceRef invalidFQCellInstRef(const IHEN& hen)
{
    return FQCellInstanceRef(hen.invalidCellRef(), hen.invalidCellInstanceRef());
}

} // unnamed namespace
Utils::HierNetGraphProxyVertexProperties::HierNetGraphProxyVertexProperties(const IHEN& hen, const CellInstData& cellInstPropValues, const InterfaceType& p)
{
    typedef PromotedVertexPropertyObtainer PropObtainerType;
    this->setPropertyManager(p);
    this->initPropertyTable(PropObtainerType(hen, cellInstPropValues, p));
}

Utils::HierNetGraphProxyEdgeProperties::HierNetGraphProxyEdgeProperties(const IHEN& hen, const CellInstData& cellInstPropValues, const InterfaceType& p)
{
    typedef PromotedEdgePropertyObtainer PropObtainerType;
    this->setPropertyManager(p);
    this->initPropertyTable(PropObtainerType(hen, cellInstPropValues, p));
}

Utils::FlatNetIterator
Utils::getFlatNets(const IHEN& hen, CellRef cell, IFlatNetFilter* netFilter)
{
    PONTE_TRACE_BLOCK(40, HEN, "getFlatNets: " << cell );
    using namespace Iterators;

    return makePolymorphIt<FlatNetIteratorImpl>(hen, cell, netFilter);
}

CellRef
Utils::getTopCell(const IHEN& hen)
{
        IHEN::CellIterator cellIt = hen.getCells();
        if ( hen.cellCount() == 1 )
    {
        va_check( cellIt != null(), "" );
        return *cellIt;
    }

    // iterate over all cells and find the only cell that has no instantiations

    typedef Map<CellRef, size_t> CellInstantiationCount;
    CellInstantiationCount cellInstCount;
    // fill the cell instantiation map
    for ( ; cellIt != null(); ++cellIt )
    {
        const CellRef parentCell = *cellIt;
        for( IHEN::CellInstanceIterator subCellIt = hen.getSubCells(parentCell); subCellIt != null(); ++subCellIt )
        {
            const CellRef subCell = hen.getCellInstPrototype(parentCell / *subCellIt);
            ++cellInstCount[subCell];
        }
    }

    CellRef topCellRef = hen.invalidCellRef();
    // look for cells that are not in the map
    for ( cellIt = hen.getCells(); cellIt != null(); ++cellIt )
    {
        const CellRef cell = *cellIt;
        if ( !contains(cellInstCount, cell) )
        {
            PONTE_ASSUMPTION(topCellRef == hen.invalidCellRef(), "There must be a single top cell");
            topCellRef = cell;
        }
    }

    PONTE_ASSERT(topCellRef != hen.invalidCellRef(), "");
    return topCellRef;
}

CellRef
Utils::getCellByName(const IHEN& hen, const std::string& cellName)
{
    typedef IHEN::CellProperties::ConstProperty<std::string> CellNameProp;
    const CellNameProp cellNameProp(&hen.cellProperties(), CELL_NAME_PROP);

    CellRef cellRef = hen.invalidCellRef();
    // look for cells that are not in the map
    for ( IHEN::CellIterator cellIt = hen.getCells(); cellIt != null(); ++cellIt )
    {
        const CellRef cell = *cellIt;
        const std::string& thisCellName = cellNameProp.getPropertyValue(cell);
        if ( thisCellName == cellName )
        {
            PONTE_ASSUMPTION(cellRef == hen.invalidCellRef(), "Cell names must be unique");
            cellRef = cell;
        }
    }

    PONTE_ENSURE(cellRef != IHEN::invalidCellRef(), ("HEN_CELL_NOT_FOUND", cellName));

    return cellRef;
}

RectangleType
Utils::graphBoundingBox(IHEN::ConstNetGraphPtr graphPtr)
{
    typedef IHEN::IGraphVertexReadOnlyProperties ReadOnlyNodeProps;
    typedef ReadOnlyNodeProps::ConstProperty<CoordPropType> ConstNodeCoordProp;

    const ReadOnlyNodeProps& vertexProps = *(graphPtr.as<const ReadOnlyNodeProps>());

    // TODO: use IPropertySet
    const ConstNodeCoordProp xProp = vertexProps.getProperty<CoordPropType>(PG_NODE_X_INT_PROPERTY);
    const ConstNodeCoordProp yProp = vertexProps.getProperty<CoordPropType>(PG_NODE_Y_INT_PROPERTY);

    RectangleType boundingBox;
    GraphLib::IReadOnlyGraph::VertexIterator vertexIt = graphPtr.as<const GraphLib::IReadOnlyGraph>()->vertices();
    for ( ; vertexIt != null(); ++vertexIt )
    {
        const CoordPropType x = xProp.getPropertyValue(*vertexIt);
        const CoordPropType y = yProp.getPropertyValue(*vertexIt);

        PONTE_ASSERT(propertyValueIsValid(x) && propertyValueIsValid(y), "");

        if (boundingBox.isValid())
            RectangleType::Utl::enlarge(boundingBox, PointType(x, y));
        else
            boundingBox = RectangleType(PointType(x, y), PointType(x, y));
    }

    return boundingBox;
}

std::string
getNetName(const IHEN& hen, const FQNetRef& net)
{
    typedef IHEN::NetProperties::ConstProperty<std::string> NetStrProp;
    using PredefinedProps::NET_NAME_PROP;

    const NetStrProp netNameProp(&hen.netProperties(), NET_NAME_PROP);
    return netNameProp.getPropertyValue(net);
}

double
getNetCapacitance(const IHEN& hen, const FQNetRef& net)
{
    typedef IHEN::NetProperties NetPropMgr;
    typedef NetPropMgr::ConstProperty<double> NetDoubleProp;
    using PredefinedProps::NET_CAPACITANCE_PROP;

    const NetPropMgr& netPropMgr = hen.netProperties();
    const NetDoubleProp netNameProp(&netPropMgr, NET_CAPACITANCE_PROP);
    return netNameProp.getPropertyValue(net);
}

FQCellInstanceRef
Utils::getFQCellInstByName(const IHEN& hen,
                           const CellRef parentCell,
                           const std::string& subCellName)
{
    typedef IHEN::CellInstProperties CellInstPropMgr;
    typedef CellInstPropMgr::ConstProperty<std::string> CellInstNameProp;
    const CellInstNameProp cellInstNameProp (&hen.cellInstProperties(), CELLINST_NAME_PROP);

    FQCellInstanceRef fqCellInstRef = invalidFQCellInstRef(hen);

    typedef IHEN::CellInstanceIterator CellInstIterator;
    PONTE_FOREACH(const CellInstIterator::ValueType& subCell, hen.getSubCells(parentCell))
    {
        const FQCellInstanceRef fqSubCell (parentCell, subCell);
        const std::string name = cellInstNameProp.getPropertyValue(fqSubCell);

        if (name == subCellName)
        {
            PONTE_ASSUMPTION(fqCellInstRef == invalidFQCellInstRef(hen),
                             "Sub cell instance names must be unique");

            fqCellInstRef = fqSubCell;
        }
    }
    return fqCellInstRef;
}

std::string
Utils::getCellInstanceName(const IHEN& hen,
                           const FQCellInstanceRef& inst)
{

    typedef IHEN::CellInstProperties CellInstPropMgr;
    typedef CellInstPropMgr::ConstProperty<std::string> CellInstNameProp;

    const CellInstPropMgr& cellInstPropMgr = hen.cellInstProperties();
    const CellInstNameProp cellInstNameProp =
        cellInstPropMgr.getProperty<std::string>(PredefinedProps::CELLINST_NAME_PROP);

    return cellInstNameProp.getPropertyValue(inst);
}

} // namespace HEN